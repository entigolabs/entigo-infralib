## Opinionated helm package for crossplane ##

This module works togeather with "modules/aws/eks" - it provides the needed SA and AWS Role when __crossplane_enable__ is true.

This will also initialize the [AWS crossplane provider](https://github.com/crossplane-contrib/provider-aws/releases).

The helm package is made up of 3 argocd sync waves. For module testing with only pure helm installProvider and installProviderConfig booleans are combined.



### Example code ###

```
    modules:
      - name: crossplane-system
        source: crossplane

```
### Limitations ###
Currently the module has to use the name "crossplane-system" or it will not function correctly. Only one installation per Kubernetes cluster.
