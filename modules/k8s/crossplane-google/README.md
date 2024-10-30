## Opinionated helm package for crossplane ##

This module depends on: modules/k8s/crossplane-core

This will initialize the [provider-gcp](https://github.com/crossplane-contrib/provider-gcp).

The Helm package is made up of 2 ArgoCD sync waves.



### Example code ###

```
    modules:
      - name: crossplane-k8s
        source: crossplane-k8s

```

