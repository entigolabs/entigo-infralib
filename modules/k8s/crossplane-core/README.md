## Opinionated helm package for crossplane ##

This module depends on: modules/aws/crossplane or modules/google/crossplane

This will initialize [crossplane](https://github.com/crossplane/crossplane).


### Example code ###

```
    modules:
        - name: crossplane-system
          source: crossplane-core

```

### Limitations ###
Currently the module has to use the name "crossplane-system" or it will not function correctly. Only one installation per Kubernetes cluster.
