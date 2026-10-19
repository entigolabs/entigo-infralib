## Opinionated helm package for Application Gateway for Containers ##

Gateway API gateways on Application Gateway for Containers (AGC).
The vendored alb-controller creates the AGC in the AKS node resource group from the `ApplicationLoadBalancer`.

Depends on: modules/azure/aks (`alb_controller_client_id`, the chart needs the client id at render time),
modules/azure/vpc (`agc_subnets`), modules/k8s/cert-manager (AGC reads certificates only from Kubernetes Secrets).
The subscription needs the `Microsoft.ServiceNetworking` and `Microsoft.NetworkFunction` resource providers.

* One `Gateway` (one AGC frontend with its own `*.fzXX.alb.azure.com` FQDN, no static IP) per enabled `gateways` entry,
  named `<release name>-<key>`: HTTPS with a cert-manager wildcard for `<domain>`, HTTP only redirects (`sslRedirect`).
* Internal: AGC has no private frontend yet (Azure/AKS#5739). istio-gateway creates the Gateway
  `global.internalGateway` (`<release name>-internal`) in `internalGatewayNamespace`. Consumer charts chain
  `.toptin.azure-gateway.global.internalGateway` and `.toptin.azure-gateway.global.internalGatewayNamespace`.
* One ALB controller per cluster. A second azure-gateway module sets `installAlbController: false` and either its own
  `ApplicationLoadBalancer` in another /24 of vpc `agc_subnets` (`global.azure.agcSubnetIds` in the config inputs, the
  agent default is the first subnet) or `applicationLoadBalancer.create: false` with `name`/`namespace` of the first
  module's one (one AGC per /24). `agcSubnetIds` is a list because the `associations` field is one; Azure accepts a
  single association per AGC today.

## SSL policies

`sslPolicy` per gateway takes the AGC predefined policy name. Default
`2023-06` (TLS 1.2 + 1.3), `2023-06-S` is the stricter variant, see values.yaml. AGC has no TLS 1.0/1.1 policy.

### Example code ###

```
    modules:
      - name: azure-gateway
        source: azure-gateway
```
