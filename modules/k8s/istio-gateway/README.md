## Opinionated helm package for istio-gateway ##
This will create an istio gateway object combined with an AWS ALB Ingress or Google gateway object.

This module is used to expose istio to outside of a Kubernetes cluster.
A separate VirtualService object has to be created to expose the application.


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
