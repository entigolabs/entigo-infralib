## Crossplane identities ##

Creates two user-assigned identities with federated credentials for the AKS OIDC issuer:

* `<prefix>` for the Crossplane Azure providers (ServiceAccount `crossplane-system/crossplane-azure`):
  Contributor and Role Based Access Control Administrator on the resource group, so any provider or composition works
  without extra permissions here. Role assignments are allowed for all roles except Owner, User Access Administrator
  and Role Based Access Control Administrator. On the AKS node resource group it may only assign Monitoring Reader
  (grafana Azure Monitor datasource).
* `<prefix>-core` for Crossplane core (ServiceAccount `crossplane-system/crossplane`): AcrPull on the acr-proxy
  registry (`acr_id`) to pull the packages.

Needs azure/aks (`oidc_issuer_url`, `node_resource_group_id`) in the same step or an earlier one, and acr-proxy
in an earlier step (its id decides what is created).

The federated credentials trust fixed service accounts: `kubernetes_namespace` (`crossplane-system`) with
`kubernetes_service_account` (`crossplane-azure`, the k8s crossplane-azure providers) and
`kubernetes_core_service_account` (`crossplane`, k8s crossplane-core). Renaming those modules or their namespace
breaks the Azure login without an obvious error: change these inputs to match.
