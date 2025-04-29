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
        version: stable

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
