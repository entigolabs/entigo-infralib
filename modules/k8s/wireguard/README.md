## Helm charts that we use

These modules can be used in the [entigo-infralib-agent](https://github.com/entigolabs/entigo-infralib-agent) steps of "**type: argocd-apps**".

## Example code

```
steps:
  - name: apps
    type: argocd-apps
    modules:
      - name: wireguard
        source: wireguard

```


dns_policy_vpc_ids must be set in google/dns to resolve DNS names from private DNS zones

WireGuard client configuration example for Google Cloud:

```
[Interface]
PrivateKey = <your super secret private key>
Address = 172.31.200.2/32 # IP address assigned to your device in Google VPC
DNS = 10.149.128.25 # Required to resolve DNS names from private DNS zones (gcloud compute addresses list --filter DNS_RESOLVER).
MTU = 1380

[Peer]
PublicKey = VPvvKbhsmQx0jK9KROKmVQGUSH25Re5xwe9R+MI7hz8= # Public key of the WireGuard server. Can be obtained from WireGuard server logs. (kubectl logs <wireguard-pod>)
AllowedIPs = 10.149.0.0/16, 172.0.0.0/8 # IP-s routed through WireGuard. Add Google VPC, GKE Control Plane CIDR, private service access CIDR etc.
Endpoint = 35.228.101.151:51820 # WireGuard server endpoint. (kubectl get service <wireguard-service>)
PersistentKeepalive = 15
```

## Oracle ##

Two things differ from aws and google, both because of the same OCI limitation.

**The load balancer is a different product.** WireGuard is UDP, and the OCI *classic* Load
Balancer - the one `oci-native-ingress-controller` drives - carries only TCP and HTTP. A UDP
Service has to become a **Network** Load Balancer, which the cloud-controller-manager selects
from `oci.oraclecloud.com/load-balancer-type: "nlb"`. Without that annotation the Service stays
`Pending` indefinitely and emits nothing useful.

The NLB is also invisible to the VCN's security lists, and the CCM will edit those itself unless
told not to. So `security-list-management-mode: "None"` keeps it out and
`oci.oraclecloud.com/oci-network-security-groups` points at `modules/oracle/oke`'s
`nlb_nsg_id`, which opens UDP 51820. Get that wrong and the load balancer reports healthy while
dropping every handshake.

`is-preserve-source: "true"` matters more here than it looks: without it every packet arrives
with the load balancer's address and no handshake can be attributed to a peer.

**The public key is served through the ingress, not the load balancer.** On aws the key rides on
the NLB's port 443 with `aws-load-balancer-ssl-cert`. An OCI network load balancer is pure layer
4 with no TLS termination, so there is no certificate to attach - `templates/oracle/pubkey.yaml`
puts a Service and an Ingress on the ordinary ingress load balancer instead, which is closer to
what google does with an HTTPRoute.

The pubkey host is on the **public** zone deliberately: it is the one thing that must be
reachable before the VPN is up.

### Peers ###

`clients` is per-deployment - set it in the deployment's own config, never here. Generate a
keypair locally and publish only the public half:

```shell
wg genkey | tee wg-private.key | wg pubkey
```

### Reaching a private zone through it ###

The point of the VPN on Oracle is the private DNS zone, which resolves only inside the VCN. A
client therefore needs both the routes and the resolver:

```ini
[Interface]
PrivateKey = <your wg-private.key>
Address    = 172.31.201.2/32
DNS        = <kube-dns ClusterIP>  # CoreDNS, which forwards to the VCN resolver

[Peer]
PublicKey  = <fetched from the pubkey host>
Endpoint   = wireguard.<public domain>:51820
AllowedIPs = 172.31.201.0/24, 10.156.0.0/16, 10.96.0.0/16
```

`10.156.0.0/16` is the VCN (where the internal load balancer lives) and `10.96.0.0/16` the
services CIDR, which is what makes CoreDNS reachable - and CoreDNS is what resolves a private
zone, since it forwards to the VCN resolver. Substitute your own CIDRs: `services_cidr` in
`modules/oracle/oke` and the VCN's own block.

**Read the DNS address, do not derive it.** OKE does not place kube-dns at `.10` of the
services CIDR the way many clusters do - on a `10.96.0.0/16` cluster it was `10.96.5.5`. Get it
from the cluster:

```shell
kubectl get svc -n kube-system kube-dns -o jsonpath='{.spec.clusterIP}'
```

**The Endpoint hostname resolves to two addresses.** An OCI network load balancer has both a
public and a private IP, the Service reports both, and external-dns publishes both under the
one name. WireGuard resolves the hostname once at startup and picks one - so roughly half the
time it picks the in-VCN address and the tunnel silently never establishes. Until that is
fixed, put the public IP in `Endpoint` rather than the hostname.
