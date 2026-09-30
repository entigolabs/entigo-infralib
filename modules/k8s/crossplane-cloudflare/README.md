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

