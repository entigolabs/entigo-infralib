## Bring-your-own Application Gateway for Containers with private frontends ##

One private Application Gateway for Containers (AGC) per module for k8s/azure-gateway: `agc-internal` in vpc default
mode, `agc-control` and `agc-service` in spoke mode (examples below). The public gateway stays on the AGC that the ALB controller manages itself
(k8s/azure-gateway `ApplicationLoadBalancer`); the ALB controller can't put a private frontend on a managed AGC yet
([Azure/AKS#5739](https://github.com/Azure/AKS/issues/5739)), but it configures bring-your-own ones with the
`azure-alb-internal` GatewayClass (`privateFrontend`).

* AGC `<prefix>` in the agent resource group, associated with `subnet_id`: a /24 delegated to
  `Microsoft.ServiceNetworking/trafficControllers` that no other AGC uses (CNI Overlay: exactly a /24, one AGC per
  subnet). Default the vpc private-agc subnet (`private_agc_subnets[0]`); in spoke mode the control-agc subnet (internal
  gateway) and the service-agc subnet (service gateway), further AGCs a vpc custom subnet.
* Private frontend `private` (`publicNetworkAccess: Disabled`) with an FQDN
  `<id>.privatelink.alb.azure.com`. azurerm has no private frontend, so it is an azapi resource on the GA API
  2026-03-01.
* The private frontend is a Private Link resource: clients reach it through the private endpoint `<prefix>` in
  `private_endpoint_subnet_id` (default the vpc node subnet `private_subnets[0]`; in spoke mode the modules set
  `control_pe_subnets[0]` and `service_pe_subnets[0]`). Its
  zone group writes the FQDN's A record into `private_dns_zone_id` (default the dns module's `privatelink.alb.azure.com`,
  a hub zone with dns `agc_private_dns_zone_id`).
* The ALB controller identity (aks `alb_controller_principal_id`, the aks module needs `agc_subnet_ids`) gets AppGw for
  Containers Configuration Manager on the AGC and Network Contributor on the subnet (CNI Overlay extension config).
* Outputs for k8s/azure-gateway: `agc_id` (Gateway annotation `alb.networking.azure.io/alb-id`) and `frontend_name`
  (Gateway address `alb.networking.azure.io/alb-frontend`), plus `frontend_fqdn`, `private_endpoint_ip` and
  `subnet_cidr` (source of the AGC traffic to the pods, for NetworkPolicies).

Every module is a separate load balancer in its own subnet, in spoke mode with its own private endpoint subnet, like the
aws spoke ALBs (control, service). k8s/azure-gateway wires its `internal` gateway from `.toutput.agc`
(the only agc module, or the one with `default_module: true`) and its `service` gateway from the module named
`agc-service` in step `infra`; further gateways get `{{ .output.<step>.<module>.agc_id }}` and `frontend_name` in its
inputs. Cost: about 134 USD a month
per AGC (AGC, association and frontend hours, list price) plus capacity units and the private endpoint.

### Example code ###
Default mode, on the vpc private-agc subnet:
```
      - name: agc-internal
        source: azure/agc
```
Spoke mode, a control and a service gateway (k8s/azure-gateway spoke example for the gateways):
```
      - name: agc-control
        source: azure/agc
        default_module: true
        inputs:
          subnet_id: "{{ .toutput.vpc.control_agc_subnets[0] }}"
          private_endpoint_subnet_id: "{{ .toutput.vpc.control_pe_subnets[0] }}"
      - name: agc-service
        source: azure/agc
        inputs:
          subnet_id: "{{ .toutput.vpc.service_agc_subnets[0] }}"
          private_endpoint_subnet_id: "{{ .toutput.vpc.service_pe_subnets[0] }}"
```
