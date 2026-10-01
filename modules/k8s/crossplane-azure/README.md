## Opinionated helm package for crossplane on Azure ##

This module depends on: modules/k8s/crossplane-core, modules/azure/crossplane, modules/azure/aks

Installs the [provider-azure](https://github.com/crossplane-contrib/provider-upjet-azure) family (`global.requiredProviders`,
mirrored with `pullpush.sh`). Providers use the workload identity of modules/azure/crossplane, federated with
`crossplane-system/crossplane-azure`: keep the module name `crossplane-azure`. One per cluster.

### WorkloadIdentity ###

Creates the managed identity `<cluster name>-<namespace>-<service account>`, its federated credential, the role
assignments and the service account with the client id (charts set `serviceAccount.create: false`), plus
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

Roles: `workloadIdentity.allowedRoles`, only on resources in the agent resource group, storage roles only on a blob
container other than `tfstate`. Whoever can create a WorkloadIdentity in a namespace can grant these roles.

### Example code ###

```
    modules:
      - name: crossplane-azure
        source: crossplane-azure
```
