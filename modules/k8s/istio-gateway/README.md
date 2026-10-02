## Opinionated helm package for istio-gateway ##
This will create an istio gateway object combined with an AWS ALB Ingress or Google gateway object.

This module is used to expose istio to outside of a Kubernetes cluster.
A separate VirtualService object has to be created to expose the application.


### Azure ###

* Public, like google: istio `Gateway` (port 80) for VirtualServices, an HTTPRoute for `*.<pub_domain>` on the
  azure-gateway external gateway (AGC) and an AGC HealthCheckPolicy (15021 `/healthz/ready`).
* Internal, until AGC has a private frontend (Azure/AKS#5739): only in the module whose namespace is azure-gateway
  `internalGatewayNamespace`. Service `<gateway name>-internal` (internal load balancer) and Gateway API gateway
  `<azure-gateway module>-internal` (sectionName `https`) with a cert-manager wildcard for `int_domain`, records from
  the external-dns private instance. With several istio-gateway modules, set azure-gateway
  `global.internalGatewayNamespace` (or `default_module`).

Depends on: azure-gateway, cert-manager, istio-istiod.

### Example code ###

```
    modules:
        - name: generic-gw-ext
          source: istio-gateway
          inputs:
            gateway:
                name: generic-gw-ext
            global:
              aws:
                certificateArn: '{{ .toutput.route53.pub_cert_arn }}'
                groupName: internal
                scheme: internal
        - name: generic-gw-int
          source: istio-gateway
          inputs:
            gateway:
                name: generic-gw-int
            global:
              aws:
                certificateArn: '{{ .toutput.route53.int_cert_arn }}'
                groupName: internal
                scheme: internal

```
