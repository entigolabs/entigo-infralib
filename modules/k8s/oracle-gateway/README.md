# oracle-gateway

The shared ingress edge for OKE: one Kubernetes Gateway API `Gateway`, served by Istio
(`gatewayClassName: istio`), fronted by `modules/k8s/oci-native-ingress-controller` (NIC)
instead of a Kubernetes `LoadBalancer` Service. Plays the same role for Oracle that
`modules/k8s/google-gateway` plays for GKE.

Requires, in this order: `gateway-api-crds` (wave 0), `istio-base` + `istio-istiod` (wave
1), `oci-ingress` (wave 2, its `IngressClass` has to exist first), then this module (wave
3).

## Why NIC fronts this instead of a LoadBalancer Service

A Gateway API `Gateway` normally gets its own `LoadBalancer` Service, provisioned by the
Kubernetes Cloud Controller Manager (CCM). That was tried here first and reverted: the
CCM's only TLS annotation (`service.beta.kubernetes.io/oci-load-balancer-tls-secret`) takes
a Kubernetes Secret, not an OCID - and the *only* way to get a Terraform-issued OCI
Certificate's actual PEM content (rather than a reference to it) is to shell out to the
`oci` CLI, since the `oracle/oci` Terraform provider has no data source for it. That's a
real, working fix (an `external` data source), but it's also the first CLI-shellout
anywhere in this codebase, and genuinely novel machinery for what used to be one annotation.

Fronting with NIC instead sidesteps the whole problem: NIC already terminates client TLS
from a plain certificate OCID (`templates/ingress.yaml`) - the same way every app's
`Ingress` already worked before this migration - and Istio's Gateway just becomes NIC's
backend, exactly as an app's own pods were before. Nothing here needs the wildcard
certificate's actual PEM content at all.

## The `Ingress` and the `Gateway` are two separate TLS sessions

`templates/ingress.yaml` is a wildcard-host `Ingress` using the real wildcard certificate
by OCID - the client-facing TLS session, same as NIC has always worked.
`templates/gateway.yaml`'s `Gateway` listener is a *second*, independent TLS session
between NIC and Envoy, using a throwaway self-signed certificate (`templates/secret.yaml`)
that exists only so this hop is encrypted rather than plaintext.

That second certificate's chain of trust does not matter, and is not checked: traced
through NIC's actual production source
(`pkg/loadbalancer/loadbalancer.go`/`pkg/controllers/ingress/util.go` in
`oracle/oci-native-ingress-controller`), its backend TLS support only ever sets
`TrustedCertificateAuthorityIds` on the OCI load balancer's backend set - it never sets
`VerifyPeerCertificate`, which the OCI SDK defaults to `false` when unset. So
`backend-tls-enabled` (left at its default, `true`, unlike every app's `Ingress` before it)
genuinely produces encrypted-but-unverified TLS today, not full verification, regardless of
what certificate or CA is attached. The actual protection against something impersonating
Envoy on this hop is OCI's own network fabric, not this certificate: OCI binds each
instance's expected IP address to the physical port it is connected to, with a reverse-path
check against encapsulation tampering, so nothing else on the VCN can spoof its way onto
this path in the first place.

## Internal vs. external: two instances of this chart

Each needs its own OCI load balancer and its own `IngressClass`, so each is its own Helm
release - the same `oci-ic`/`oci-ic-int` split every app's `Ingress` already had:

```yaml
modules:
  - name: oracle-gateway
    source: oracle-gateway
    # public: uses agent_input_oracle.yaml's defaults (oci-ic, pub_cert_ocid) as-is
  - name: oracle-gateway-int
    source: oracle-gateway
    inputs:
      global:
        oracle:
          ingressClassName: oci-ic-int
          certificateOcid: "{{ .toutput.dns.int_cert_ocid }}"
```

istiod (the control plane) stays a single shared install regardless of how many of these
are deployed - only the data-plane Envoy workload duplicates.

## How apps attach

Apps use a Gateway API `HTTPRoute` naming this Gateway (and, for the internal instance,
its own namespace) via `parentRefs`:

```yaml
parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: oracle-gateway-int
    namespace: oracle-gateway-int
    sectionName: https
```

There is only one listener, so `sectionName` is not load-bearing the way it would be with
multiple listeners - set anyway for clarity and so a future second listener can't silently
start matching routes that never asked for it.

## Gotchas worth knowing

- **The generated Service must stay `ClusterIP`.** Istio defaults a Gateway's Service to
  `LoadBalancer`, which the CCM would then provision into a second, unwanted, billed OCI
  load balancer - `networking.istio.io/service-type: ClusterIP` on the `Gateway`'s own
  `metadata.annotations` is what stops that.
- **An Istio gateway pod does not listen on a port until a listener binds it.** The OCI
  load balancer's backend stays unhealthy until this Gateway exists and is Programmed.
- **The backend certificate churns on every Helm render unless reused.** `genSelfSignedCert`
  produces fresh output every time, so `templates/secret.yaml` looks up the existing Secret
  and reuses its key material across upgrades - same pattern, same reason, as
  `oci-native-ingress-controller`'s own webhook certificate. `argo-apps.yaml` still ignores
  this Secret's `/data`, since a first sync (nothing to look up yet) renders fresh anyway.
