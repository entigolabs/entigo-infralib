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

### Provider resources ###

All provider pods share the requests and limits in `providerResources.default` (`providerResources.family` for `upbound-provider-family-aws`). To size one subpackage differently, add it under `providerResources.providers`; the given fields are deep-merged over `default` and that provider gets its own DeploymentRuntimeConfig. Which providers are installed is still decided by `global.requiredProviders`, `global.extraProviders` and the observer settings.

```
providerResources:
  providers:
    ec2:
      requests:
        cpu: 150m
        memory: 384Mi
      limits:
        memory: 1536Mi
```
