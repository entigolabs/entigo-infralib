## OCI Native Ingress Controller ##

Deploys Oracle's [OCI Native Ingress Controller](https://github.com/oracle/oci-native-ingress-controller)
(NIC) and a default `IngressClass`/`IngressClassParameters` pair backed by a public OCI
Load Balancer, so other modules (e.g. `argocd`) can get a real hostname via a standard
`Ingress` resource.

### Why this module vendors its own chart copy ###

Unlike every other `modules/k8s/*` module, this one does **not** use a `dependencies:`
entry + `Chart.lock` pointing at a Helm chart repository - NIC has no chart repository,
only an in-repo chart at `github.com/oracle/oci-native-ingress-controller/helm/oci-native-ingress-controller`.
`charts/oci-native-ingress-controller/` is a manual copy of that chart at tag `v1.4.3`.
To upgrade: diff the upstream `helm/oci-native-ingress-controller` directory at the new
tag against this one and re-apply the two deviations below.

### Deviations from the upstream chart ###

1. **`templates/webhook.yaml` was dropped** (not copied). Upstream's version creates a
   `cert-manager.io` `Certificate`/`Issuer` for the webhook's TLS cert - this repo has no
   `cert-manager` module. Instead, `templates/webhook-certs.yaml` (in this module, not the
   vendored subchart) generates an equivalent self-signed CA/cert directly via Helm's
   `genCA`/`genSignedCert`, stored as the same `oci-native-ingress-controller-tls` Secret
   the vendored `deployment.yaml` already expects, with the CA inlined directly as the
   `MutatingWebhookConfiguration`'s `caBundle` instead of relying on cert-manager's
   `cert-manager.io/inject-ca-from` annotation.
   - This is **not optional/cosmetic**: `main.go` unconditionally starts a webhook HTTPS
     server at boot (`pkg/server/server.go`'s `SetupWebhookServer`) and calls `os.Exit(1)`
     if it can't load a cert from `/tmp/k8s-webhook-server/serving-certs` - the controller
     will crash-loop without *some* cert there, even though the webhook's own feature
     (pod-readiness-gate injection) is opt-in per namespace and unused by default.
   - `fullnameOverride: oci-native-ingress-controller` is set in this module's
     `values.yaml` specifically so the generated Secret name is deterministic and known
     ahead of time to `webhook-certs.yaml` - Helm subchart named templates aren't reliably
     callable from a parent chart's own templates across chart boundaries.
2. **`ingressClassParameters.subnetId`/`.compartmentId` and
   `oci-native-ingress-controller.subnet_id`** are wired from
   `.toutput.vpc.public_subnet_id`, a new **singular** output added to `modules/oracle/vpc`
   alongside the existing `public_subnets` list output - the agent's k8s Helm value
   templating (`agent_input_oracle.yaml`) has no way to index into a list output the way
   Terraform module variables get auto-wired, so anything needing exactly one OCID as a
   plain string needs its own singular output.

### Example code ###

```
    modules:
      - name: oci-ingress
        source: oci-native-ingress-controller
```
