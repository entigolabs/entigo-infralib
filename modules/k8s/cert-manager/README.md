## Opinionated helm package for cert-manager ##

cert-manager with an ACME `ClusterIssuer` (DNS-01 in the cloud DNS), default `letsencrypt`. One per cluster.

Azure only for now (AzurePublicCloud): Application Gateway for Containers reads certificates only from Kubernetes
Secrets (aws and google gateways use ACM / Certificate Manager). The controller identity is a WorkloadIdentity
(crossplane-azure) with DNS Zone Contributor on the dns `pub_zone_id` and `int_cert_zone_id` zones. Depends on:
modules/azure/dns, modules/k8s/crossplane-azure.

Another cloud adds `values-<cloud>.yaml` (`global.cloudProvider`) and `templates/<cloud>/` with its ClusterIssuer
(the azure one with its DNS-01 solver) and DNS identity.

Test environments should use `acme.server: https://acme-staging-v02.api.letsencrypt.org/directory`: production
allows 5 certificates per identical name set per 7 days.

### Issuers ###

* `clusterIssuer.name` and `acme.server` select the ACME CA. CAs that need External Account Binding (ZeroSSL, Sectigo,
  Google Public CA): `acme.externalAccountBinding` with a Secret holding the HMAC key in this namespace.
* `extraClusterIssuers`: more ClusterIssuers with the spec as is (CA, Vault, ...).
* azure-gateway and istio-gateway default to `clusterIssuer.name` of this module, their `certificateIssuer` picks
  another one.

### Example code ###

```
    modules:
      - name: cert-manager
        source: cert-manager
```
