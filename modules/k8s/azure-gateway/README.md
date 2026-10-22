## Opinionated helm package for Application Gateway for Containers ##

Gateway API gateways on Application Gateway for Containers (AGC).
The vendored alb-controller creates the AGC in the AKS node resource group from the `ApplicationLoadBalancer`.

Depends on: modules/azure/aks (`alb_controller_client_id`, the chart needs the client id at render time),
modules/azure/vpc (`public_agc_subnets`), modules/azure/agc (the internal gateway's AGC), modules/k8s/cert-manager (AGC reads
certificates only from Kubernetes Secrets).
The subscription needs the `Microsoft.ServiceNetworking` and `Microsoft.NetworkFunction` resource providers.

* One `Gateway` (one AGC frontend with its own `*.fzXX.alb.azure.com` FQDN, no static IP) per enabled `gateways` entry,
  named after its key (`external`): HTTPS with a cert-manager wildcard for `<domain>`, HTTP only redirects (`sslRedirect`).
* Internal: the managed AGC can't have a private frontend yet (Azure/AKS#5739), so the Gateway `internal`
  (`global.internalGateway`, in this module's namespace) is on the bring-your-own AGC of the azure/agc module:
  GatewayClass `azure-alb-internal` (alb-controller `privateFrontend: true`), annotation
  `alb.networking.azure.io/alb-id: <agc_id>`, address `alb.networking.azure.io/alb-frontend: <frontend_name>`, HTTPS with
  a cert-manager wildcard for the int domain. The agent sets both (`annotations`, `addresses`) from `.toutput.agc` (the
  only agc module, or the one with `default_module: true`); gateways without `addresses` are on the managed AGC. Clients resolve the route hostnames through external-dns CNAMEs to
  the frontend's `*.privatelink.alb.azure.com` FQDN, which only resolves where the dns module's `privatelink.alb.azure.com`
  zone (or the hub's) is linked. Consumer charts chain `.toptin.azure-gateway.global.internalGateway` and
  `.toptin.azure-gateway.global.internalGatewayNamespace`; their backends get an AGC `HealthCheckPolicy` (every route
  on Azure is on AGC).
* More private gateways, e.g. the spoke service gateway: one azure/agc module each and one `gateways` entry like
  `internal` (example below).
* One azure-gateway module per cluster (it installs the ALB controller): more gateways are more `gateways` entries.

## SSL policies

`sslPolicy` per gateway takes the AGC predefined policy name. Default
`2023-06` (TLS 1.2 + 1.3), `2023-06-S` is the stricter variant, see values.yaml. AGC has no TLS 1.0/1.1 policy.

### Example code ###

```
    modules:
      - name: azure-gateway
        source: azure-gateway
```
Spoke mode with a service gateway (azure/agc modules `agc-control` with `default_module: true` and `agc-service` in
step `infra`):
```
      - name: azure-gateway
        source: azure-gateway
        inputs:
          gateways:
            service:
              enabled: true
              gatewayClassName: azure-alb-internal
              domain: "service.{{ .toutput.dns.int_domain }}"
              sslRedirect: true
              annotations:
                alb.networking.azure.io/alb-id: "{{ .output.infra.agc-service.agc_id }}"
              addresses:
                - type: alb.networking.azure.io/alb-frontend
                  value: "{{ .output.infra.agc-service.frontend_name }}"
```
