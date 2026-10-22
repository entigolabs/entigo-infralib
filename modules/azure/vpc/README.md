## Opinionated module for Azure VNet creation ##

Creates a VNet with public, private, intra and database subnets, a delegated public-agc subnet for the managed public
Application Gateway for Containers (AGC), a delegated private-agc subnet for the internal gateway's AGC, a delegated
aks subnet (API Server VNet Integration), a delegated Container Apps subnet with its environment (the agent pipeline
jobs), optional custom subnets and a NAT gateway. Spoke mode replaces private-agc with control-agc and service-agc
subnets for private AGCs and control-pe and service-pe subnets for their private endpoints.

### Automatic subnet calculations ###
If only `vpc_cidr` is given, it is split in quarters:

* first quarter: public and intra next to each other (a /25 each, a /26 each for /20 and smaller), the rest is spare
* second quarter: private
* third quarter: database (an eighth) and the delegated block (an eighth): public-agc, private-agc, aks and
  container-apps, the rest is spare
* fourth quarter: spare, never used by the split, for `custom_subnets` and explicit extra subnets

A /21 and a /22 use a small layout: public, intra, aks and container-apps in the first /24, public-agc in the second,
private in the second half. private-agc takes the third /24 in a /20 and a /21; a /22 has no room for it.

Delegated subnets:

* public-agc: a /24 (AGC with Azure CNI Overlay needs exactly a /24, one AGC per subnet, up to 5 frontends per AGC).
* private-agc: a /24 like public-agc, the AGC of the internal gateway (azure/agc `subnet_id` defaults to it). Automatic
  in default mode only.
* aks: a /26 (a /27 for /20 and smaller), 59 or 27 usable IPs. A cluster reserves at least 9, so clusters can share
  it, e.g. for a blue/green replacement. For a larger or extra subnet, set `aks_subnets`.
* container-apps: a /26 (a /27 for /20 and smaller), delegated to `Microsoft.App/environments`. The module creates an
  internal Container Apps environment (Consumption profile, zone redundant by default with
  `container_apps_zone_redundancy_enabled`, create time only) in each Container Apps subnet: it runs the agent pipeline
  jobs of steps with `vpc: attach`, and the agent finds it by subnet and waits until it is ready (10+ minutes). One
  environment per subnet, minimum /27, 14 IPs reserved: a /26 runs about 250 Consumption replicas, a /27 about 90. The
  size can't change once an environment uses it. Infrastructure resource group `<subnet name>-<location>`, job logs in
  the `<prefix>-container-apps` Log Analytics workspace.

Database subnets are for private endpoints and are never delegated. Default mode creates one for /19 and larger
(smaller VNets need explicit `database_subnets`); spoke mode always creates one.

Any subnet list can be set explicitly; the split doesn't shift around it. Extra Container
Apps environments are extra `container_apps_subnets` entries, extra AGCs are custom subnets.

Minimum size for the automatic split: /22 in default mode, /21 in spoke mode. A /23 or /24 needs every list set explicitly (`[]` for
unused tiers, see the example below); a /24 has no room for an AGC, a /23 in spoke mode none for anything beyond the
two AGC /24s.

| vpc_cidr | public | intra | private | database | public-agc | private-agc | aks | container apps | spare |
|---|---|---|---|---|---|---|---|---|---|
| /16 | /25 | /25 | /18 | /19 | /24 | /24 | /26 | /26 | /18 + /19 + /20 + /20 + /21 + /21 + /22 + /22 + /23 + /24 + /24 + /25 |
| /17 | /25 | /25 | /19 | /20 | /24 | /24 | /26 | /26 | /19 + /20 + /21 + /21 + /22 + /22 + /23 + /24 + /24 + /25 |
| /18 | /25 | /25 | /20 | /21 | /24 | /24 | /26 | /26 | /20 + /21 + /22 + /22 + /23 + /24 + /24 + /25 |
| /19 | /25 | /25 | /21 | /22 | /24 | /24 | /26 | /26 | /21 + /22 + /23 + /24 + /24 + /25 |
| /20 | /26 | /26 | /22 | none | /24 | /24 | /27 | /27 | /22 + /23 + /24 + /24 + /25 + /25 + /26 |
| /21 | /26 | /26 | /22 | none | /24 | /24 | /27 | /27 | /24 + /26 |
| /22 | /26 | /26 | /23 | none | /24 | none | /27 | /27 | /26 |

### subnet_split_mode ###
`default`:

| Subnet | Use | Outputs |
|---|---|---|
| public-0 | resources with a public IP or public load balancer | `public_subnets` |
| intra-0 | no internet access, e.g. attachments and appliances | `intra_subnets` |
| private-0 | nodes, internal load balancers (`type: LoadBalancer` Services) and the internal gateway's private endpoint | `private_subnets`, `compute_subnets` |
| database-0 (/19 and larger) | private endpoints of databases and caches | `database_subnets` |
| public-agc-0 (/24) | AGC of the public gateway (managed by the ALB controller) | `public_agc_subnets` |
| private-agc-0 (/24) | AGC of the internal gateway (azure/agc) | `private_agc_subnets` |
| aks-0 | AKS API Server VNet Integration | `aks_subnets` |
| container-apps-0 | Container Apps environment (agent pipeline jobs) | `container_apps_subnets` |

`spoke` is for a VNet peered to a hub that owns the firewall, gateways and DNS and writes its rules per role range:

| Subnet | Use | Outputs |
|---|---|---|
| public-0 | resources with a public IP or public load balancer | `public_subnets` |
| compute-0 | nodes and internal load balancers (the private subnet renamed) | `private_subnets`, `compute_subnets` |
| database-0 | private endpoints of databases and caches | `database_subnets` |
| public-agc-0 (/24) | AGC of the public gateway (managed by the ALB controller) | `public_agc_subnets` |
| control-agc-0 (/24) | AGC of the internal gateway (azure/agc) | `control_agc_subnets`, `control_agc_subnet_names` |
| service-agc-0 (/24) | AGC of the service gateway (azure/agc) | `service_agc_subnets`, `service_agc_subnet_names` |
| control-pe-0 (/28) | private endpoint of the internal gateway (the clients' entry IP) | `control_pe_subnets` |
| service-pe-0 (/28) | private endpoint of the service gateway | `service_pe_subnets` |
| aks-0 | AKS API Server VNet Integration | `aks_subnets` |
| container-apps-0 | Container Apps environment (agent pipeline jobs) | `container_apps_subnets` |

* control-agc and service-agc are delegated to `Microsoft.ServiceNetworking/trafficControllers` and hold only their
  AGC (a separate load balancer each, like the aws spoke ALBs). They take the first two /24s of the VNet's second
  eighth (in a /21 after public-agc).
* control-pe and service-pe have private endpoint network policies on, so NSGs and route tables apply. They take the
  first two /28s after public (in a /21 after container-apps).
* No AGC or private endpoint subnet has a NAT gateway or egress route.
* No intra subnet (the peering attaches the VNet) and always a database subnet (a /23 for /20, a /26 in the intra slot
  for /21).
* No private-agc (control-agc takes its role); public, database (/19 and larger), public-agc, aks and container-apps
  keep their default-mode addresses.
* The four lists (`control_agc_subnets`, `service_agc_subnets`, `control_pe_subnets`, `service_pe_subnets`, names via
  `*_subnet_names`) can be set explicitly in either mode; spoke mode needs at least one control-agc and one
  service-agc subnet. Their outputs are empty in default mode unless set.

| vpc_cidr | public | compute | database | public-agc | aks | container apps | control-agc | service-agc | control-pe | service-pe | spare |
|---|---|---|---|---|---|---|---|---|---|---|---|
| /16 | /25 | /18 | /19 | /24 | /26 | /26 | /24 | /24 | /28 | /28 | /18 + /20 + /20 + /20 + /21 + /21 + /21 + /22 + /22 + /22 + /23 + /23 + /24 + /24 + /24 + /25 + /26 + /27 |
| /17 | /25 | /19 | /20 | /24 | /26 | /26 | /24 | /24 | /28 | /28 | /19 + /21 + /21 + /21 + /22 + /22 + /22 + /23 + /23 + /24 + /24 + /24 + /25 + /26 + /27 |
| /18 | /25 | /20 | /21 | /24 | /26 | /26 | /24 | /24 | /28 | /28 | /20 + /22 + /22 + /22 + /23 + /23 + /24 + /24 + /24 + /25 + /26 + /27 |
| /19 | /25 | /21 | /22 | /24 | /26 | /26 | /24 | /24 | /28 | /28 | /21 + /23 + /23 + /24 + /24 + /24 + /25 + /26 + /27 |
| /20 | /26 | /22 | /23 | /24 | /27 | /27 | /24 | /24 | /28 | /28 | /22 + /24 + /25 + /25 + /26 + /27 |
| /21 | /26 | /22 | /26 | /24 | /27 | /27 | /24 | /24 | /28 | /28 | /27 |

Switching an existing VNet to spoke drops private-agc and intra (in a /21 intra must be empty: database takes its
range) and renames the node subnet, which Azure does by recreating it (aks then replaces the cluster): pin the old
name with `private_subnet_names` when a cluster runs.

### Custom subnets ###
`custom_subnets` is a map keyed by a short name (1-40 lowercase letters or digits), read by the agent as
`{{ .toutput.vpc.custom_subnets[key] }}` (ids) and `custom_subnet_cidrs[key]`. Fields:

* `cidr`: usually from the spare quarter
* `name`: default `<prefix>-<key>`, e.g. `AzureBastionSubnet` for fixed names
* `delegation`, `delegation_actions` (default the join action)
* `network_security_group`, `route_table`: an empty NSG / route table the delegated service manages
* `nat_gateway`: internet egress like the private subnets (not with `route_table`); without it the subnet has none

Delegated services take a custom subnet, never a database subnet: SQL Managed Instance (delegation with the network
policies actions, NSG and route table, no NAT gateway, minimum /27; a /24 fits 24 Business Critical instances) or a
VNet-integrated Postgres/MySQL flexible server. Further AGCs (more azure/agc modules) take an exact /24 delegated to
`Microsoft.ServiceNetworking/trafficControllers`, for /19 and larger e.g. from the spare part of the delegated block.

```
          custom_subnets:
            mssql:
              cidr: "10.156.224.0/27"
              delegation: Microsoft.Sql/managedInstances
              delegation_actions:
                - Microsoft.Network/virtualNetworks/subnets/join/action
                - Microsoft.Network/virtualNetworks/subnets/prepareNetworkPolicies/action
                - Microsoft.Network/virtualNetworks/subnets/unprepareNetworkPolicies/action
              network_security_group: true
              route_table: true
            bastion:
              name: AzureBastionSubnet
              cidr: "10.156.224.64/26"
            agcextra:
              cidr: "10.156.163.0/24"
              delegation: Microsoft.ServiceNetworking/trafficControllers
```

### Egress ###
Default outbound access is off on all subnets. The private and Container Apps subnets egress through the NAT gateway
(`nat_static_ip_count` public IPs, about 64k SNAT ports each; `nat_gateway_sku` StandardV2 zone redundant or Standard,
changing it replaces the gateway and its IPs; `nat_idle_timeout_minutes`). Intra and database subnets have no internet
access, public subnets only through a resource's own public IP or load balancer. To egress through a firewall/NVA
(e.g. in a peered hub) set `enable_nat_gateway: false`, `egress_next_hop_ip` (0.0.0.0/0 route on the private and
Container Apps subnets) and aks `outbound_type: userDefinedRouting`; without either, AKS and pipeline jobs fail.

The private and Container Apps subnets have Storage and Key Vault service endpoints (`*_subnet_service_endpoints`):
these services see the private IP instead of the NAT IP, so firewalls need VNet rules.

### Names ###
Subnet names default to `<prefix>-<type>-<n>`, override with `*_subnet_names`. Container Apps subnet names are also
the environment names (lowercase letters, digits, hyphens, max 60). Removing or reordering Container Apps subnets
recreates the environments after the removed one; the agent finds them again on the next run.

### Flow logs ###
`enable_flow_log` (default true) sends all VNet flows to a dedicated storage account `<prefix alnum, max 14>fl<unique
suffix>` (Entra ID only, Microsoft-managed keys, network access for trusted Azure services and the private subnets), or
to an existing account in the same region (`flow_log_storage_account_id`). Retention `flow_log_retention_days` (7).
The flow log lives in the region's Network Watcher (`NetworkWatcherRG`/`NetworkWatcher_<location>`, override with
`network_watcher_name`/`network_watcher_resource_group_name`) and is deleted with the VNet.
`flow_log_traffic_analytics_enabled` adds Traffic Analytics into a `<prefix>-flow-log` Log Analytics workspace (per GB).

* Microsoft-managed keys, because rotating a customer-managed key stops flow logs silently until they are re-enabled
  ([Self-managed key rotation](https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#storage-account)).
* Not logged: Container Apps, SQL Managed Instance and Postgres/MySQL flexible server traffic
  ([incompatible services](https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#incompatible-services)).
* Cost: per GB collected, 5 GB/month free per subscription, plus storage
  ([pricing](https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#pricing)).

### Plan time checks ###
`vpc_cidr` IPv4 and /24 or larger, `subnet_split_mode`, overlaps, subnets inside the VNet, AGC subnets exactly /24,
unique names, `custom_subnets` keys and `nat_gateway`/`route_table`, `egress_next_hop_ip` only without NAT gateway,
explicit lists below /22 (spoke: /21), at least one control-agc and service-agc subnet in spoke mode, `unique_suffix`.
Azure itself requires aks /28 or larger and Container Apps /27 or larger.

### Outputs ###
* Per tier `*_subnets` (ids) and `*_subnet_cidrs`: private, public, intra, database, public_agc, private_agc, aks,
  container_apps; `container_apps_environment_ids`; `control_agc_subnets`, `service_agc_subnets`, `control_pe_subnets`,
  `service_pe_subnets` (+ `_cidr_blocks`); `custom_subnets`, `custom_subnet_cidrs`; `vpc_id`, `vpc_name`, `vpc_cidr`.
* Names: `private/public/intra/database_subnet_names`, `control_agc_subnet_names`, `service_agc_subnet_names`.
* Cross-cloud names: `name`, `vpc_cidr_block`, `*_subnets_cidr_blocks`, `nat_public_ips`, `compute_subnets` (the private
  subnets), `elasticache_subnets_cidr_blocks` (database).
* `nat_static_ips`; `nat_gateway_id`, `egress_route_table_id` and `vpc_flow_log_*` are empty strings when disabled, so
  agent inputs referencing them always render.

Not supported yet: IPv6 / dual stack.

### Examples ###
```
    modules:
      - name: vpc
        source: azure/vpc
        inputs:
          vpc_cidr: "10.156.0.0/16"
```

Default mode, `10.156.0.0/16` down to `/22`:

| Subnet | /16 | /17 | /18 | /19 | /20 | /21 | /22 |
|---|---|---|---|---|---|---|---|
| public-0 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/26 | 10.156.0.0/26 | 10.156.0.0/26 |
| intra-0 | 10.156.0.128/25 | 10.156.0.128/25 | 10.156.0.128/25 | 10.156.0.128/25 | 10.156.0.64/26 | 10.156.0.64/26 | 10.156.0.64/26 |
| private-0 | 10.156.64.0/18 | 10.156.32.0/19 | 10.156.16.0/20 | 10.156.8.0/21 | 10.156.4.0/22 | 10.156.4.0/22 | 10.156.2.0/23 |
| database-0 | 10.156.128.0/19 | 10.156.64.0/20 | 10.156.32.0/21 | 10.156.16.0/22 | none | none | none |
| public-agc-0 | 10.156.160.0/24 | 10.156.80.0/24 | 10.156.40.0/24 | 10.156.20.0/24 | 10.156.10.0/24 | 10.156.1.0/24 | 10.156.1.0/24 |
| private-agc-0 | 10.156.161.0/24 | 10.156.81.0/24 | 10.156.41.0/24 | 10.156.21.0/24 | 10.156.2.0/24 | 10.156.2.0/24 | none |
| aks-0 | 10.156.162.0/26 | 10.156.82.0/26 | 10.156.42.0/26 | 10.156.22.0/26 | 10.156.11.0/27 | 10.156.0.128/27 | 10.156.0.128/27 |
| container-apps-0 | 10.156.162.64/26 | 10.156.82.64/26 | 10.156.42.64/26 | 10.156.22.64/26 | 10.156.11.32/27 | 10.156.0.160/27 | 10.156.0.160/27 |
| spare | 10.156.1.0-10.156.63.255, 10.156.162.128-10.156.255.255 | 10.156.1.0-10.156.31.255, 10.156.82.128-10.156.127.255 | 10.156.1.0-10.156.15.255, 10.156.42.128-10.156.63.255 | 10.156.1.0-10.156.7.255, 10.156.22.128-10.156.31.255 | 10.156.0.128-10.156.1.255, 10.156.3.0/24, 10.156.8.0/23, 10.156.11.64-10.156.15.255 | 10.156.0.192/26, 10.156.3.0/24 | 10.156.0.192/26 |

Spoke mode (`subnet_split_mode: spoke`), `10.156.0.0/16` down to `/21`:

| Subnet | /16 | /17 | /18 | /19 | /20 | /21 |
|---|---|---|---|---|---|---|
| public-0 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/26 | 10.156.0.0/26 |
| control-pe-0 | 10.156.0.128/28 | 10.156.0.128/28 | 10.156.0.128/28 | 10.156.0.128/28 | 10.156.0.64/28 | 10.156.0.192/28 |
| service-pe-0 | 10.156.0.144/28 | 10.156.0.144/28 | 10.156.0.144/28 | 10.156.0.144/28 | 10.156.0.80/28 | 10.156.0.208/28 |
| control-agc-0 | 10.156.32.0/24 | 10.156.16.0/24 | 10.156.8.0/24 | 10.156.4.0/24 | 10.156.2.0/24 | 10.156.2.0/24 |
| service-agc-0 | 10.156.33.0/24 | 10.156.17.0/24 | 10.156.9.0/24 | 10.156.5.0/24 | 10.156.3.0/24 | 10.156.3.0/24 |
| compute-0 | 10.156.64.0/18 | 10.156.32.0/19 | 10.156.16.0/20 | 10.156.8.0/21 | 10.156.4.0/22 | 10.156.4.0/22 |
| database-0 | 10.156.128.0/19 | 10.156.64.0/20 | 10.156.32.0/21 | 10.156.16.0/22 | 10.156.8.0/23 | 10.156.0.64/26 |
| public-agc-0 | 10.156.160.0/24 | 10.156.80.0/24 | 10.156.40.0/24 | 10.156.20.0/24 | 10.156.10.0/24 | 10.156.1.0/24 |
| aks-0 | 10.156.162.0/26 | 10.156.82.0/26 | 10.156.42.0/26 | 10.156.22.0/26 | 10.156.11.0/27 | 10.156.0.128/27 |
| container-apps-0 | 10.156.162.64/26 | 10.156.82.64/26 | 10.156.42.64/26 | 10.156.22.64/26 | 10.156.11.32/27 | 10.156.0.160/27 |
| spare | 10.156.0.160-10.156.31.255, 10.156.34.0-10.156.63.255, 10.156.161.0/24, 10.156.162.128-10.156.255.255 | 10.156.0.160-10.156.15.255, 10.156.18.0-10.156.31.255, 10.156.81.0/24, 10.156.82.128-10.156.127.255 | 10.156.0.160-10.156.7.255, 10.156.10.0-10.156.15.255, 10.156.41.0/24, 10.156.42.128-10.156.63.255 | 10.156.0.160-10.156.3.255, 10.156.6.0/23, 10.156.21.0/24, 10.156.22.128-10.156.31.255 | 10.156.0.96-10.156.1.255, 10.156.11.64-10.156.15.255 | 10.156.0.224/27 |

A /23 or /24 needs every list explicit, e.g. `10.156.0.0/23` without Container Apps:
```
    modules:
      - name: vpc
        source: azure/vpc
        inputs:
          vpc_cidr: "10.156.0.0/23"
          public_subnets: '["10.156.0.0/27"]'
          intra_subnets: '["10.156.0.32/27"]'
          aks_subnets: '["10.156.0.64/27"]'
          container_apps_subnets: '[]'
          private_subnets: '["10.156.0.128/25"]'
          database_subnets: '[]'
          public_agc_subnets: '["10.156.1.0/24"]'
```
