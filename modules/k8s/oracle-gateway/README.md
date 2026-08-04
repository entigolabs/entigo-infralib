# oracle-gateway

The shared ingress edge for OKE: one Kubernetes Gateway API `Gateway`, served by Istio
(`gatewayClassName: istio`), behind an OCI load balancer that does nothing but pass TCP
80/443 through. TLS terminates in the Istio gateway with full SNI. Plays the same role
for Oracle that `modules/k8s/google-gateway` plays for GKE.

Requires, in this order: `gateway-api-crds` (wave 0), `istio-base` + `istio-istiod` and
`cert-manager` (wave 1), then this module (wave 2).

## Why the Gateway API and not Istio's own Gateway CRD

Istio's `networking.istio.io/Gateway` reads its TLS material from `tls.credentialName`,
and those secrets **must** live in the namespace of the gateway *workload* - cross-namespace
references are never permitted and there is no ReferenceGrant equivalent. Per-app
certificates would therefore all have to be created in the gateway's namespace, or copied
there by a secret reflector.

With the Gateway API, cert-manager's gateway-shim creates each certificate's Secret in the
Gateway's own namespace, which is exactly where Istio runs the generated proxy - so the
two agree by construction, with nothing to copy.

The other decisive reason is ACME. cert-manager's classic HTTP-01 solver creates an
`Ingress`, and with ingress-nginx gone nothing would serve it; Istio's plain-Ingress
support is its most legacy surface and is cluster-global via `meshConfig.ingressSelector`.
The Gateway API path uses `http01.gatewayHTTPRoute` instead, which needs no ingress
controller at all.

## Listeners and how apps attach

- One shared `http` listener on port 80 - HTTP→HTTPS redirect, ACME HTTP-01 challenges,
  and OCI load balancer health checking. `allowedRoutes.namespaces.from: All`, because
  cert-manager creates its solver routes in the issuer's namespace.
- One `https-<name>` listener on port 443 per entry in `hosts`, each with
  `hostname: <name>.<domain>` and its own `<name>-tls` certificate. Envoy picks between
  them by SNI. Multiple listeners on the same port are how the Gateway API expresses SNI
  multiplexing.

Apps attach with an `HTTPRoute` naming the listener explicitly:

```yaml
parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: oracle-gateway
    namespace: oracle-gateway
    sectionName: https-argocd
```

`sectionName` is not optional in practice. Without it the route also attaches to the
port-80 listener (whose hostname is unset, so it matches everything) and the app would be
served over plaintext HTTP alongside HTTPS.

`hosts` entries must match the app module names, since the listener name and the
certificate secret name are both derived from them. Set the list per environment:

```yaml
- name: oracle-gateway
  source: oracle-gateway
  inputs:
    hosts:
      - argocd
      - hello-world
```

Adding an app means adding it here as well as deploying it - the coupling is deliberate
and mirrors `google-gateway`, where the shared edge owns the listeners and apps only
attach routes.

## Per-host certificates, not a wildcard

A single wildcard certificate would collapse all of this into one listener, but Let's
Encrypt only issues wildcards through DNS-01, and cert-manager has no built-in OCI DNS
solver (it would need a webhook). HTTP-01 cannot issue wildcards. Per-host certificates
also isolate failures - one app's certificate problem doesn't take the others down - and
avoid Istio's `IST0138` warning about HTTP/2 connection reuse across hosts sharing a
certificate.

## Gotchas worth knowing

- **An Istio gateway pod does not listen on a port until a listener binds it.** Unlike
  nginx, which always listens on 80/443, the OCI load balancer's backends stay unhealthy
  until this Gateway exists. The Gateway is load-bearing for LB health, not just routing.
- **`infrastructure.annotations` replaces rather than merges** the annotations Istio would
  otherwise copy from the Gateway's own `metadata.annotations`. All OCI load balancer
  settings must live under `service.annotations` (which renders into `infrastructure`),
  never on the Gateway metadata.
- **A listener with no `hostname`, or `tls.mode` other than `Terminate`, is silently
  skipped by cert-manager** - no Certificate, no event, no error.
- **The redirect and ACME coexist safely.** cert-manager's solver route matches the
  challenge path with `type: Exact`, which Gateway API ranks above this module's
  `PathPrefix: /` redirect rule.
