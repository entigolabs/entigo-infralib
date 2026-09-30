## Crossplane identities ##

Creates two user-assigned identities with federated credentials for the AKS OIDC issuer:

* `<prefix>` for the Crossplane Azure providers (ServiceAccount `crossplane-system/crossplane-azure`):
  Contributor and Role Based Access Control Administrator on the resource group, so any provider or composition works
  without extra permissions here. Role assignments are allowed for all roles except Owner, User Access Administrator
  and Role Based Access Control Administrator.
* `<prefix>-core` for Crossplane core (ServiceAccount `crossplane-system/crossplane`): AcrPull on the acr-proxy
  registry (`acr_id`) to pull the packages.

Needs azure/aks (`oidc_issuer_url`) in the same step or an earlier one.
