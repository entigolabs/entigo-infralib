## Opinionated helm package for crossplane on Azure ##

This module depends on: modules/k8s/crossplane-core, modules/azure/crossplane, modules/azure/aks

Installs the [provider-azure](https://github.com/crossplane-contrib/provider-upjet-azure) family (`global.requiredProviders`,
mirrored with `pullpush.sh`). Providers use the workload identity of modules/azure/crossplane, federated with
`crossplane-system/crossplane-azure`: keep the module name `crossplane-azure`. One per cluster.

Namespaced managed resources only (`*.azure.m.upbound.io`) with
`providerConfigRef: {kind: ClusterProviderConfig, name: crossplane-azure}`; there is no cluster-scoped `ProviderConfig`.

### WorkloadIdentity ###

Creates the managed identity `<cluster name>_<namespace>_<name>` (Kubernetes names have no `_`, so names can't collide
across namespaces), its federated credential, the role assignments (deterministic names) and the service account with
the client id (charts set `serviceAccount.create: false`), plus
`serviceAccountAnnotations` if set.

```
apiVersion: azure.entigo.com/v1alpha1
kind: WorkloadIdentity
metadata:
  name: external-dns
spec:
  serviceAccountName: external-dns
  roleAssignments:
    - role: DNS Zone Contributor
      scope: /subscriptions/<id>/resourceGroups/<agent rg>/providers/Microsoft.Network/dnsZones/example.com
```

Roles: `workloadIdentity.roles` maps each role that a WorkloadIdentity may grant to the scopes it may grant it on
(regexes over the whole lower case scope, `{rg}` = agent resource group, `{nodeRg}` = AKS node resource group,
`{encryptionKeyIds}` = one of `workloadIdentity.encryptionKeyIds`). By default: storage roles only on a blob
container (not `tfstate`), DNS roles only on a zone, AcrPull only on a registry, Key Vault Secrets User only on a vault
or secret, all in the agent resource group, Monitoring Reader also on the resource groups themselves and the AKS node
resource group. Roles not listed, or whose patterns are all unset, are rejected. Scopes must be plain resource ids:
Azure resolves dot segments and percent-encoding, the rules compare strings. Whoever can create a WorkloadIdentity in
a namespace can grant these roles. XR names can't contain dots (not allowed in Azure identity names).
The same role and scope can't be listed twice (case-insensitive). `serviceAccountAnnotations` can't set
`azure.workload.identity/client-id` or `gotemplating.fn.crossplane.io/*`, the composition sets them.

Without `serviceAccountName` only the identity and its role assignments are created (no federated credential, no
service account), e.g. for a storage account customer managed key (loki/mimir `<release>-cmk`). `serviceAccountName`
can't be added, removed or changed later. Key Vault Crypto Service Encryption User is only allowed without
`serviceAccountName` and only on the keys in `workloadIdentity.encryptionKeyIds` (kms telemetry key).

`keepOnDelete: true` keeps the Azure identity, federated credential and role assignments when the WorkloadIdentity is
deleted (no `Delete` management policy); a new WorkloadIdentity with the same name adopts them again (external names
are deterministic). loki/mimir use it for the CMK identity, so an uninstall keeps the storage account readable.
It also applies to a role assignment removed from `roleAssignments` (or a changed scope): it stays in Azure and has to
be removed by hand, Crossplane can't tell a removed entry from a deleted WorkloadIdentity.

The composition needs `global.azure.subscriptionID`, `resourceGroupName`, `location`, `oidcIssuerUrl` and
`identityPrefix` (agent inputs), the chart fails to render without them.

### Example code ###

```
    modules:
      - name: crossplane-azure
        source: crossplane-azure
```
