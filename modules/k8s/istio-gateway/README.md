## Opinionated helm package for istio-gateway ##
This will create an istio gateway object combined with an AWS ALB Ingress object.

This module is used to expose istio to outside of a Kubernetes cluster.
You need to specify the domain names to use in the inputs. This should result in an AWS ALB being provisioned and also external-dns should configure the DNS accordingly.
A separate VirtualService object has to be created to expose the application.

__ingressInboundAllow__ can be used to specify incoming IP range that is allowed, defaults to "0.0.0.0/0" (allow all).



### Example code ###

```
    modules:
      - name: istio-gateway
        source: istio-gateway
        inputs:
          domain: "*.ep-infrastructure-dns-test.infralib.entigo.io"
          altdomain:
            - "*.something.ep-infrastructure-dns-test.infralib.entigo.io"

```

### Limitations ###
This module needs development. Not tested properly.
