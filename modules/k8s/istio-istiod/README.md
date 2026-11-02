## Opinionated helm package for istio-istiod ##
This will install istio runtime into your cluster. No additional values need to be specified.

This module is needed if you want to use istio and is combined with istio-base and istio-gateway modules.

### Example code ###

```
    modules:
      - name: istio-system
        source: istio-istiod

```

### Envoy response headers ###
By default the module keeps Envoy's own `x-envoy-decorator-operation` (shows internal service names) and `x-envoy-upstream-service-time` headers out of responses to end users, with the `remove-envoy-headers` EnvoyFilter. To keep the headers:

```
    modules:
      - name: istio-system
        source: istio-istiod
        inputs:
          removeEnvoyHeaders: false
```

### Limitations ###
Currently the module has to use the name "istio-system" or it will not function correctly. Multiple installations are not supported at this point. One per Kubernetes cluster is allowed.
