# gateway-api-crds

Installs the upstream Kubernetes Gateway API CRDs (standard channel, **v1.5.1**) on
clusters that don't provide them. Cluster-scoped and cloud-agnostic - the manifest is a
verbatim upstream bundle with no templating.

Only needed where nothing else supplies the CRDs:

- **oracle (OKE)**: needed. Oracle states plainly that "the Gateway API Custom Resource
  Definitions (CRDs) are not installed by default in clusters you create using Kubernetes
  Engine", so this module is what makes `modules/k8s/oracle-gateway` (and Istio's
  `gatewayClassName: istio`) work at all.
- **google (GKE)**: not needed, GKE ships them.
- **aws (EKS)**: not needed, `modules/k8s/aws-alb` bundles its own copy of the same
  v1.5.1 manifest for the ALB gateway controller.

## Version constraint

The version is not free to pick: **Istio 1.30 requires Gateway API v1.5.x**. Istio's own
upgrade notes warn that with older CRDs "`TLSRoute` and `ReferenceGrant` resources will
become invisible to istiod" - a silent failure that shows up as a Gateway with
`attachedRoutes: 0` and no Envoy listener, not as an error. Upstream Gateway API has
since released v1.6, but Istio 1.30's docs still pin v1.5.1, so don't bump this ahead of
the istio-* modules without checking Istio's compatibility notes first.

Keep this bundle and `modules/k8s/aws-alb/templates/gateway-api-crds.yaml` on the same
version - they are the same file, and letting them drift means the two clouds disagree
about which Gateway API they support.

## Ordering

Deployed at `infralib.entigo.io/sync-wave: "0"`, ahead of cert-manager and Istio, because
both only look for these CRDs when their pods start:

- cert-manager performs the Gateway API check on startup only, so CRDs arriving later
  leave `config.gatewayAPI.enabled` doing nothing until the deployment is restarted
  (`kubectl rollout restart deployment cert-manager -n cert-manager`).
- istiod decides what to watch at startup too.

This matters when retrofitting an existing cluster; on a fresh provision wave 0 already
guarantees the right order.
