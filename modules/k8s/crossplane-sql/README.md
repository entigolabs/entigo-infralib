## Opinionated helm package for crossplane ##

This module depends on: modules/k8s/crossplane-core

This will initialize the [provider-sql](https://github.com/crossplane-contrib/provider-sql).

The Helm package is made up of 2 ArgoCD sync waves.

### Example code ###

```yaml
    modules:
      - name: crossplane-sql
        source: crossplane-sql

```
