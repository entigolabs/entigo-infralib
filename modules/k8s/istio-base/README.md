## Opinionated helm package for istio-base ##
This will install istio CRDs into your cluster. No additional values need to be specified.

This module is needed if you want to use istio and is combined with istio-istiod and istio-gateway.

### Example code ###

```
    modules:
      - name: istio-base
        source: istio-base

```

### Limitations ###
Multiple installations are not supported at this point. One per Kubernetes cluster is allowed.
