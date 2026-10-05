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

Roles: `workloadIdentity.allowedRoles`, only on resources in the agent resource group, with the scope shape per role
from `workloadIdentity.scopePatterns` (storage roles only on a blob container other than `tfstate`, DNS roles only on a
zone, ...). Scopes must be plain resource ids: Azure resolves dot segments and percent-encoding, the rules compare
strings. `workloadIdentity.readOnlyRoles` (Monitoring Reader) also on the agent resource group
itself and the AKS node resource group. Whoever can create a WorkloadIdentity in a namespace can grant these roles.

Without `serviceAccountName` only the identity and its role assignments are created (no federated credential, no
service account), e.g. for a storage account customer managed key (loki/mimir `<release>-cmk`). `serviceAccountName`
can't be added, removed or changed later. Key Vault Crypto Service Encryption User is only allowed without
`serviceAccountName` and only on the keys in `workloadIdentity.encryptionKeyIds` (kms telemetry key).

`keepOnDelete: true` keeps the Azure identity, federated credential and role assignments when the WorkloadIdentity is
deleted (no `Delete` management policy); a new WorkloadIdentity with the same name adopts them again (external names
are deterministic). loki/mimir use it for the CMK identity, so an uninstall keeps the storage account readable.

### Example code ###

```
    modules:
      - name: crossplane-azure
        source: crossplane-azure
```
