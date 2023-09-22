## Opinionated helm package for istio-istiod ##
This will install istio runtime into your cluster. No additional values need to be specified.

This module is needed if you want to use istio and is combined with istio-base and istio-gateway modules.

### Example code ###

```
    modules:
      - name: istio-system
        source: istio-istiod

```

### Limitations ###
Currently the module has to use the name "istio-system" or it will not function correctly. Multiple installations are not supported at this point. One per Kubernetes cluster is allowed.
