## Opinionated helm package for crossplane on Azure ##

This module depends on: modules/k8s/crossplane-core, modules/azure/crossplane, modules/azure/aks

Installs the [provider-azure](https://github.com/crossplane-contrib/provider-upjet-azure) family (`global.requiredProviders`,
mirrored with `pullpush.sh`). Providers use the workload identity of modules/azure/crossplane, federated with
`crossplane-system/crossplane-azure`: keep the module name `crossplane-azure`. One per cluster.

Namespaced managed resources only (`*.azure.m.upbound.io`) with
`providerConfigRef: {kind: ClusterProviderConfig, name: crossplane-azure}`; there is no cluster-scoped `ProviderConfig`.

### WorkloadIdentity ###

`iam.entigo.com/v1alpha1` WorkloadIdentity (the API shared by all clouds; Azure uses `azurePermissions`) comes from the
chart dependency [entigo-infralib-apis](https://github.com/entigolabs/entigo-infralib-apis) (values under
`entigo-infralib-apis`): the Function `entigo-infralib-apis-iam-fn`, the Configuration
`entigo-infralib-apis-workloadidentity` (XRD + Composition) and the EnvironmentConfig `entigo-infralib-apis-iam` with
the cluster settings (`global.azure`) and the resolved role allowlist. API, naming and validation:
[compositions/workloadidentity](https://github.com/entigolabs/entigo-infralib-apis/tree/main/compositions/workloadidentity).

Creates the managed identity `<cluster name>_<namespace>_<name>` (`..._nosa` without a service account, at most 128
characters), its federated credential,
the role assignments (deterministic names) and the service account with the client id (charts set
`serviceAccount.create: false`).

```
apiVersion: iam.entigo.com/v1alpha1
kind: WorkloadIdentity
metadata:
  name: external-dns
spec:
  serviceAccountName: external-dns
  azurePermissions:
    - role: DNS Zone Contributor
      scope: /subscriptions/<id>/resourceGroups/<agent rg>/providers/Microsoft.Network/dnsZones/example.com
```

Roles: `entigo-infralib-apis.iam.environmentConfig.azure.roles` maps each role a WorkloadIdentity may grant to the
scopes it may grant it on (regexes over the whole lower case scope, `{rg}` = agent resource group, `{nodeRg}` = AKS
node resource group, `{encryptionKeyIds}` = one of `encryptionKeyIds`, here the kms telemetry key). By default: storage
roles only on a blob container (not `tfstate`), DNS roles only on a zone, AcrPull only on a registry, Key Vault
Secrets User only on a vault or secret, all in the agent resource group, Monitoring Reader also on the resource groups
themselves and the AKS node resource group, Key Vault Crypto Service Encryption User only without
`serviceAccountName` on the telemetry key (loki/mimir `<release>-cmk`). Whoever can create a WorkloadIdentity in a
namespace can grant these roles, so creating WorkloadIdentities is a platform (administrator) permission: platform
charts create them, don't grant the kind to tenants (see the
[security section](https://github.com/entigolabs/entigo-infralib-apis/tree/main/compositions/workloadidentity#security-who-may-create-workloadidentities)).
`keepOnDelete: true` (only without a service account) keeps the Azure identity `<cluster name>_<namespace>_<name>_nosa`
and its role assignments when the WorkloadIdentity is deleted (loki/mimir CMK identities, so an uninstall keeps the
storage account readable).

The Function and Configuration have `helm.sh/resource-policy: keep` and ArgoCD `Delete=false`: removing the chart
doesn't delete the XRD and with it every WorkloadIdentity.

### Example code ###

```
    modules:
      - name: crossplane-azure
        source: crossplane-azure
```
