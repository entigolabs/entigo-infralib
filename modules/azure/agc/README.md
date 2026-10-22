## Bring-your-own Application Gateway for Containers with a private frontend ##

One Application Gateway for Containers (AGC) per module with one private frontend, for an internal Gateway API
gateway of k8s/azure-gateway. The public gateway stays on the AGC that the ALB controller manages itself
(k8s/azure-gateway `ApplicationLoadBalancer`); the ALB controller can't put a private frontend on a managed AGC yet
([Azure/AKS#5739](https://github.com/Azure/AKS/issues/5739)), but it configures bring-your-own ones with the
`azure-alb-internal` GatewayClass (`privateFrontend`).

* AGC `<prefix>` in the agent resource group, associated with `subnet_id`: a /24 delegated to
  `Microsoft.ServiceNetworking/trafficControllers` that no other AGC uses (CNI Overlay: exactly a /24, one AGC per
  subnet). In vpc spoke mode the control subnet (internal gateway) and the service subnet (service gateway), otherwise
  a vpc custom subnet.
* Private frontend `private` (`publicNetworkAccess: Disabled`) with an FQDN
  `<id>.privatelink.alb.azure.com`. azurerm has no private frontend, so it is an azapi resource on the GA API
  2026-03-01.
* The private frontend is a Private Link resource: clients reach it through the private endpoint `<prefix>` in
  `private_endpoint_subnet_id` (default vpc `control_pe_subnets[0]`: the control-pe /28 in spoke mode, the node subnet
  otherwise; the service gateway's module sets `service_pe_subnets[0]`). Its
  zone group writes the FQDN's A record into `private_dns_zone_id` (default the dns module's `privatelink.alb.azure.com`,
  a hub zone with dns `agc_private_dns_zone_id`).
* The ALB controller identity (aks `alb_controller_principal_id`, the aks module needs `agc_subnet_ids`) gets AppGw for
  Containers Configuration Manager on the AGC and Network Contributor on the subnet (CNI Overlay extension config).
* Outputs for k8s/azure-gateway: `agc_id` (Gateway annotation `alb.networking.azure.io/alb-id`) and `frontend_name`
  (Gateway address `alb.networking.azure.io/alb-frontend`), plus `frontend_fqdn`, `private_endpoint_ip` and
  `subnet_cidr` (source of the AGC traffic to the pods, for NetworkPolicies).

Every module is a separate load balancer in its own subnet, like the aws spoke ALBs (internal, service). k8s/azure-gateway
wires its `internal` gateway from `.toutput.agc` (the only agc module, or the one with `default_module: true`), further
gateways get `{{ .output.<step>.<module>.agc_id }}` and `frontend_name` in its inputs. Cost: about 134 USD a month
per AGC (AGC, association and frontend hours, list price) plus capacity units and the private endpoint.

### Example code ###
Default mode, the AGC /24 as a vpc custom subnet:
```
      - name: vpc
        source: azure/vpc
        inputs:
          custom_subnets: |
            {
              agcinternal = {
                cidr       = "10.156.162.0/24"
                delegation = "Microsoft.ServiceNetworking/trafficControllers"
              }
            }
...
      - name: agc
        source: azure/agc
        inputs:
          subnet_id: "{{ .toutput.vpc.custom_subnets[agcinternal] }}"
```
Spoke mode, an internal and a service gateway:
```
      - name: agc-control
        source: azure/agc
        default_module: true
        inputs:
          subnet_id: "{{ .toutput.vpc.control_subnets[0] }}"
      - name: agc-service
        source: azure/agc
        inputs:
          subnet_id: "{{ .toutput.vpc.service_subnets[0] }}"
          private_endpoint_subnet_id: "{{ .toutput.vpc.service_pe_subnets[0] }}"
```
