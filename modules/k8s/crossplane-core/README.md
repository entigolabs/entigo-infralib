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

`job.deleteInactiveProviderRevisions` runs a Job after every sync that deletes the Inactive ProviderRevisions left behind by a provider or Crossplane upgrade. With `job.restartOnInvalidProviderRevisions` the Job first waits for the revision roll to settle (the crossplane leader has held its lease for two minutes, every Provider is installed, every Active ProviderRevision is established and none was created in the last two minutes) and then checks that every active ManagedResourceDefinition is controlled by an Active ProviderRevision. One that is still controlled by an old revision means the new revision's package fetch was cut short (crossplane/crossplane#7817): with SafeStart the provider then stays scaled to zero and the rbac-manager roles miss API groups. In that case the Job restarts the `crossplane` deployment, waits for the rollout, and then deletes the Inactive ones; the restarted pod re-establishes every Active revision on its own, which is not waited for so the Job stays inside the agent's 600s application wait.

### Limitations ###
Currently the module has to use the name "crossplane-system" or it will not function correctly. Only one installation per Kubernetes cluster.
