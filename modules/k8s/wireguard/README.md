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

## The public key endpoint, and why AWS and google differ

The VPN itself is the same on both clouds: a `Service` of type `LoadBalancer`
carrying UDP 51820 to the wireguard container.

Alongside it the module serves the generated server public key over HTTPS, so
that clients can fetch it without someone reading it out of the pod logs. That
endpoint is the one thing this module cannot build the same way twice, and the
two clouds end up quite far apart.

### AWS: one load balancer, one hostname

`values-aws.yaml` adds a second port to the wireguard `Service` through the
chart's `service.extraPorts`, so the NLB carries both:

| Listener | Protocol | Target |
|---|---|---|
| 51820 | UDP | wireguard container, 51820 |
| 443 | TCP, TLS terminated | nginx pubkey sidecar, 8080 |

Two things make this work, and neither has a google equivalent:

- The NLB **terminates TLS**. `aws-load-balancer-ssl-ports: "443"` plus the ACM
  certificate from `agent_input_aws.yaml` mean the sidecar only ever serves
  plain HTTP on 8080.
- It is **one load balancer, so one IP, so one A record**. The VPN and the
  public key live on the same hostname,
  `<module name>.<route53 pub_domain>`, on different ports.

### Google: a route on the shared gateway, on its own hostname

Google gets the same nginx sidecar, but reaches it through an `HTTPRoute` on the
external gateway published by the `google-gateway` module, with a `Service` and
a `HealthCheckPolicy` of its own under `templates/google/`. The hostname is
`<module name>-pubkey.<dns pub_domain>`, separate from the VPN hostname.

It is worth writing down why, because the obvious question is why google does
not simply copy AWS:

1. **A passthrough load balancer terminates no TLS.** Google's external
   passthrough Network Load Balancer forwards traffic without decrypting it and
   holds no certificates — only the Application and proxy Network Load
   Balancers do. There is no `ssl-cert` equivalent to hang off the wireguard
   `Service`. The TLS for this host comes from the gateway's Certificate
   Manager map instead.
2. **The addresses cannot be shared.** The external gateway class
   `gke-l7-global-external-managed` takes a **global** address; a regional
   passthrough Network Load Balancer takes a **regional** one. They cannot be
   the same address resource, so the gateway and the wireguard load balancer
   always answer on different IPs — which is why the public key needs a
   hostname of its own rather than a second port on the VPN hostname.

Historically there was a third reason: a google `Service` of type
`LoadBalancer` could not carry TCP and UDP at once. **That one has expired.**
GKE supports mixed protocol `LoadBalancer` Services via
`spec.loadBalancerClass: networking.gke.io/l4-regional-external`, GA from
1.36.2-gke.1498000 and available external/IPv4-only from 1.34.1-gke.2190000.
Do not treat it as the blocker any more — the two reasons above are what keep
the designs apart, and neither is fixed by putting both protocols on one
`Service`.

Terminating TLS inside the nginx sidecar would collapse the difference, at the
cost of an ACME issuer and in-pod certificate renewal. cert-manager is not a
module in this repository today, so that is not on the table.

### What this means when reading the module

- `agent_input.yaml` holds everything both clouds share, including the sidecar,
  its volume and the VPN hostname annotations.
- `agent_input_aws.yaml` is only the ACM certificate.
- `agent_input_google.yaml` is only the pubkey hostname and the gateway to
  attach the route to.
- `global.gateway` and `global.pubkeyHostname` are declared in `values.yaml`
  but filled on google alone. On AWS they stay empty and nothing reads them,
  because `templates/google/` is gated on `global.cloudProvider`.

## Client configuration

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
