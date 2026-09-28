## Opinionated helm package for crossplane ##

This module depends on: modules/k8s/crossplane-core

This will initialize the [AWS crossplane provider](https://github.com/crossplane-contrib/provider-aws/releases).

The Helm package is made up of 2 ArgoCD sync waves.



### Example code ###

```
    modules:
      - name: crossplane-aws
        source: crossplane-aws

```

### Cleanup job ###

A provider version bump re-fetches every provider package at once, and on a slow registry proxy some fetches are cut short by Crossplane's reconcile deadline. The ProviderRevision then reports Healthy with a truncated object list (crossplane/crossplane#7817), never takes control of the missing ManagedResourceDefinitions, and with SafeStart the provider stays scaled to zero. The `job.deleteInactiveProviderRevisions` / `job.restartOnInvalidProviderRevisions` Job (same script as in crossplane-core) runs after the providers of this module have rolled: it waits for the roll to settle, restarts the `crossplane` deployment in `job.crossplaneNamespace` when an active ManagedResourceDefinition is not controlled by an Active ProviderRevision, and deletes the Inactive ProviderRevisions.
