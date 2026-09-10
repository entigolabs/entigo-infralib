## OCI Native Ingress Controller ##

Deploys Oracle's [OCI Native Ingress Controller](https://github.com/oracle/oci-native-ingress-controller)
(NIC) and one or more `IngressClass`/`IngressClassParameters` pairs, each backed by its own OCI
Load Balancer, so other modules (e.g. `argocd`) can get a real hostname via a standard
`Ingress` resource.

### Load balancers are chosen by IngressClass ###

OCI has no equivalent of the AWS load balancer controller's `alb.ingress.kubernetes.io/group.name`
annotation - nothing on an `Ingress` picks or creates a load balancer. The load balancer comes
from the `IngressClassParameters` that the `Ingress`'s class points at, so **more load balancers
means more IngressClasses**. `ingressClasses` in `values.yaml` is a map of them:

| entry | class name | load balancer | NSG |
|---|---|---|---|
| `public` | `oci-ic` (default class) | public subnet | `oke.lb_nsg_id` - ports from anywhere |
| `private` | `oci-ic-int` | private subnet, `isPrivate: true` | `oke.lb_int_nsg_id` - ports from the VCN CIDR |

`private` ships **disabled**, because enabling it provisions a second load balancer that is
billed whether or not an `Ingress` uses it. Both are wired with their subnet and NSG regardless,
so turning it on is one line in a deployment's config:

```yaml
ingressClasses:
  private:
    enabled: true
```

An app then selects it, in the deployment's own `config/apps/<module>.yaml`:

```yaml
ingress:
  ingressClassName: oci-ic-int
```

Add further entries for further load balancers; nothing about the map is limited to two.

**An internal load balancer is not the same as a private hostname.** The load balancer becomes
unreachable from outside the VCN, but a name on a public DNS zone still resolves for everyone
and still gets a certificate. For an app that should be invisible from outside, pair it with a
`private = true` domain in `modules/oracle/dns` - and note that external-dns needs
`--oci-zone-scope=PRIVATE` (or empty, for both) before it will publish into a private zone.

### TLS ###

NIC terminates TLS on the OCI load balancer, and **a load balancer listener can carry only
one key pair** ("customer can specify only one key pair per listener" - Oracle's own
wording). Since NIC places every `Ingress` that sets
`oci-native-ingress.oraclecloud.com/https-listener-port: "443"` on that one listener, apps
on this cluster cannot each bring their own certificate the way they do behind nginx or an
ALB. There is no SNI here.

So the platform tools share a single zone-wide **wildcard** certificate, created in the
infra step by `modules/oracle/dns` and referenced by OCID:

```yaml
annotations:
  oci-native-ingress.oraclecloud.com/certificate-ocid: ocid1.certificate.oc1...
  oci-native-ingress.oraclecloud.com/https-listener-port: "443"
  oci-native-ingress.oraclecloud.com/backend-tls-enabled: "false"
```

The OCID form is used rather than a Kubernetes TLS secret in `spec.tls` deliberately. A
secret must live in the `Ingress`'s own namespace, so a shared certificate would have to be
copied into every app namespace, and NIC would then import each copy into OCI Certificates
as a *separate* certificate - several different OCIDs competing for the single listener.
An OCID is namespace-free, so every app names the same certificate and they agree.

Note that `certificate-ocid` also makes NIC treat every host on that `Ingress` as
TLS-enabled and **ignore** `http-listener-port`, so these apps answer on 443 only.

The limit is one key pair per **listener**, not per load balancer, so an app that genuinely
needs its own certificate has two options short of Istio: give it a different
`https-listener-port` on the same load balancer, or put it on a different `IngressClass` and get
a load balancer of its own. A different port needs that port opened first - `lb_ingress_ports`
in `modules/oracle/oke`, since NIC never touches NSGs and a listener nothing can reach comes up
looking perfectly healthy.

### Gateway API CRDs ###

`templates/gateway-api-crds.yaml` is the upstream Gateway API v1.5.1 standard-channel
bundle, byte-identical to the copy in `modules/k8s/aws-alb`. OKE ships no Gateway API CRDs
at all, and NIC itself does not implement Gateway API - the bundle is here because it is
the cluster-wide prerequisite that has to land before anything that *does* implement it
(Istio), and this module is the natural owner of the ingress-layer CRDs, exactly as
`aws-alb` owns them on AWS. The CRDs carry `helm.sh/resource-policy: keep`, so removing
this module does not delete Gateways other things may still depend on.

### Why this module vendors its own chart copy ###

Unlike every other `modules/k8s/*` module, this one does **not** use a `dependencies:`
entry + `Chart.lock` pointing at a Helm chart repository - NIC has no chart repository,
only an in-repo chart at `github.com/oracle/oci-native-ingress-controller/helm/oci-native-ingress-controller`.
`charts/oci-native-ingress-controller/` is a manual copy of that chart at tag `v1.4.3`.
To upgrade: diff the upstream `helm/oci-native-ingress-controller` directory at the new
tag against this one and re-apply the three deviations below.

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
2. **`templates/deployment.yaml` sets `strategy: Recreate`.** Readiness depends on holding the
   leader-election lease, so with the chart's default RollingUpdate on a single replica the old
   pod will not terminate until the new one is Ready, and the new one cannot become Ready until
   the old one releases the lease. Every upgrade deadlocks until a pod is deleted by hand.
3. **`ingressClasses.<entry>.subnetId` and `oci-native-ingress-controller.subnet_id`** are
   wired from `.toutput.vpc.public_subnet_id` and `.toutput.vpc.private_subnet_id`, both
   **singular** outputs added to `modules/oracle/vpc` alongside its `public_subnets` /
   `private_subnets` lists - the agent's k8s Helm value templating (`agent_input_oracle.yaml`)
   has no way to index into a list output the way Terraform module variables get auto-wired,
   so anything needing exactly one OCID as a plain string needs its own singular output.

### Example code ###

```
    modules:
      - name: oci-ingress
        source: oci-native-ingress-controller
```
