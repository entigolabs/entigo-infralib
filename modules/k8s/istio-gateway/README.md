## Opinionated helm package for istio-gateway ##
This will create an istio gateway object combined with an AWS ALB Ingress or Google gateway object.

This module is used to expose istio to outside of a Kubernetes cluster.
A separate VirtualService object has to be created to expose the application.


### Azure ###

* Public, like google: istio `Gateway` (port 80) for VirtualServices, an HTTPRoute for `*.<pub_domain>` on the
  azure-gateway external gateway (AGC) and an AGC HealthCheckPolicy (15021 `/healthz/ready`).
* Own load balancer gateway (`global.azure.loadBalancer`): Service `<gateway name>-internal` or `-external` (Azure
  load balancer, `externalTrafficPolicy: Local`, `sourceRanges`) and Gateway API gateway `name` (sectionName `https`)
  with a cert-manager wildcard for `intDomain` / `pubDomain`, TLS in istio.
  * Internal (`internal: true`), until AGC has a private frontend (Azure/AKS#5739): created in the module whose
    namespace is azure-gateway `internalGatewayNamespace`, named `<azure-gateway module>-internal`, records from the
    external-dns private instance. Its frontend is placed in `global.azure.loadBalancer.internalSubnet` (vpc
    `control_subnet_name`: the control subnet in spoke mode, empty otherwise, which leaves the frontend in the node
    subnet; setting or changing it recreates the frontend). With several istio-gateway modules, set azure-gateway
    `global.internalGatewayNamespace` (or `default_module`).
  * Public (`internal: false`, `enabled: true`): istio without azure-gateway, records from the public external-dns
    instance. No AGC route (`createRoute` is on only with an azure-gateway module), no WAF.
* Without azure-gateway, the charts that route through a gateway (argocd, grafana, harbor, kiali, loki, mimir,
  prometheus, wireguard) need their gateway in the config inputs (their agent inputs require azure-gateway); harbor
  and wireguard create their AGC HealthCheckPolicy only with an azure-gateway module. Give each istio-gateway module
  its own `gateway.name`: the istio `Gateway` selects pods by label in every namespace.

Depends on: cert-manager, istio-istiod; azure-gateway for the AGC route and the internal gateway default.

Internal and public istio without azure-gateway:
```
    modules:
      - name: istio-gateway-int
        source: istio-gateway
        inputs:
          gateway:
            name: istio-gateway-int
          global:
            azure:
              loadBalancer:
                enabled: true
                name: internal
      - name: istio-gateway-ext
        source: istio-gateway
        inputs:
          gateway:
            name: istio-gateway-ext
          global:
            azure:
              loadBalancer:
                enabled: true
                internal: false
                name: external
                sourceRanges: ["0.0.0.0/0"]
      - name: harbor
        source: harbor
        inputs:
          harbor:
            expose:
              route:
                parentRefs:
                  - group: gateway.networking.k8s.io
                    kind: Gateway
                    name: external
                    namespace: istio-gateway-ext
                    sectionName: https
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
