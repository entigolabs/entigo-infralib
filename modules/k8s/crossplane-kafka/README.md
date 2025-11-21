## Opinionated helm package for crossplane ##

This module depends on: modules/k8s/crossplane-core and modules/k8s/crossplane-aws

This will initialize the [provider-kafka](https://github.com/crossplane-contrib/provider-kafka).

The Helm package is made up of 2 ArgoCD sync waves.

### Example code ###

```yaml
    modules:
      - name: crossplane-kafka
        source: crossplane-kafka

```
