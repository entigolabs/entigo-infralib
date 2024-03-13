## Opinionated helm package for istio-gateway ##
This will create an istio gateway object combined with an AWS ALB Ingress object.

This module is used to expose istio to outside of a Kubernetes cluster.
This should result in an AWS ALB being provisioned and also external-dns should configure the DNS accordingly.
A separate VirtualService object has to be created to expose the application.

__ingressInboundAllow__ can be used to specify incoming IP range that is allowed, defaults to "0.0.0.0/0" (allow all).
__certificateArn__ should be set to specify ACM certificate(s) used by ALB HTTPS listener.
__additionalIngressAnnotations__ can be used to add ALB ingress controller annotations to the ingress. 



### Example code ###

```
    modules:
      - name: istio-gateway
        source: istio-gateway
        inputs:
          certficateArn: "arn:aws:acm:us-west-2:xxxxx:certificate/xxxxxxx"
          additionalIngressAnnotations:
            alb.ingress.kubernetes.io/load-balancer-attributes: "idle_timeout.timeout_seconds=600"

```

### Limitations ###
This module needs development. Not tested properly.
