## Opinionated helm package for crossplane ##

This module depends on: modules/aws/crossplane or modules/google/crossplane

This will initialize [crossplane](https://github.com/crossplane/crossplane).


### Example code ###

```
    modules:
        - name: crossplane-system
          source: crossplane-core

```

### Cleanup job ###

`job.deleteInactiveProviderRevisions` runs a Job after every sync that deletes the Inactive ProviderRevisions left behind by a provider or Crossplane upgrade. With `job.restartOnInvalidProviderRevisions` the Job first waits for the revision roll to settle and checks every Active revision against the one it replaced. A revision whose `status.objectRefs` is shorter than its predecessor, or an Active ManagedResourceDefinition still controlled by an Inactive revision, means the package fetch was cut short (crossplane/crossplane#7817): with SafeStart the provider then stays scaled to zero and the rbac-manager roles miss API groups. In that case the Job restarts the `crossplane` deployment, waits for the revisions to become valid, and only then deletes the Inactive ones. It fails loudly instead of deleting when they do not recover. A provider release that genuinely drops resources also looks truncated to this check, so turn the flag off for that one upgrade.

### Limitations ###
Currently the module has to use the name "crossplane-system" or it will not function correctly. Only one installation per Kubernetes cluster.
