## Opinionated helm package for istio-gateway ##
This will create an istio gateway object combined with an AWS ALB Ingress or Google gateway object.

This module is used to expose istio to outside of a Kubernetes cluster.
A separate VirtualService object has to be created to expose the application.


### Azure ###

* istio `Gateway` (pod port 8080, Service port `http2` 80) for VirtualServices, an HTTPRoute for `*.<pub_domain>` on
  the azure-gateway external gateway (AGC) and an AGC HealthCheckPolicy (15021 `/healthz/ready`). VirtualService hosts
  get no own record: they resolve through the `*.<pub_domain>` wildcard to AGC.
* The istio gateway Service is ClusterIP: on Azure every entry is an AGC, the public one (azure-gateway `external`) and
  the private ones (azure-gateway `internal` on the azure/agc bring-your-own AGC). Gateway API routes on the internal
  gateway go straight to their Services, not through istio.
* Every istio-gateway module gets the route (`createRoute` follows the azure-gateway module), by default on AGC for
  `*.<pub_domain>`. For VirtualServices on the internal gateway, point a second module's route at it (`global.gateway`,
  `global.hostnames`, example below). This needs a private int domain separate from the public one, otherwise the two
  wildcard records clash; `global.createRoute: false` drops the route instead. Give each istio-gateway module its own
  `gateway.name`: the istio `Gateway` selects pods by label in every namespace.

Depends on: istio-istiod, azure-gateway (Gateway API CRDs, AGC route).

Public and internal VirtualServices:
```
    modules:
      - name: istio-gateway
        source: istio-gateway
      - name: istio-gateway-int
        source: istio-gateway
        inputs:
          gateway:
            name: istio-gateway-int
          global:
            gateway:
              name: internal
            hostnames:
              - "*.{{ .toutput.dns.int_domain }}"
```

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
