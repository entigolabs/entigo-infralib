## Opinionated module for Azure VNet creation ##

Creates a VNet with private, public, intra and database subnets, a delegated public-agc subnet for the managed public
Application Gateway for Containers (AGC), in spoke mode control-agc and service-agc subnets for the private AGCs and
control-pe and service-pe subnets for their private endpoints, a delegated aks subnet for AKS API Server VNet Integration, a delegated Container Apps subnet (the Entigo Infralib Agent
pipeline jobs), optional custom subnets (e.g. Azure SQL Managed Instance) and a NAT gateway for the private and
Container Apps subnets.

### Automatic subnet calculations ###
If only `vpc_cidr` is given, the subnets are calculated from it in quarters:

* first quarter: public and intra, the same size next to each other at the start of the VNet (a /25 each, a /26 each
  for /20 and smaller), the rest is spare
* second quarter: private
* third quarter: database (an eighth) and the delegated block (an eighth) for the service-delegated subnets below,
  with free room after them
* fourth quarter: spare, never used by the automatic split, for `custom_subnets` and explicit extra AGC or Container Apps
  subnets
* public-agc: the first /24 of the delegated block, the managed public AGC (AGC with Azure CNI Overlay needs exactly a /24, one AGC per subnet, up to
  5 frontends per AGC)
* aks: the /26 after the public-agc /24, a /27 for /20 and smaller (59 or 27 usable IPs; a cluster reserves at least 9, so several
  clusters can share it, e.g. for blue/green cluster replacement; more: set `aks_subnets`)
* container-apps: the subnet after aks, the same size (a /26, a /27 for /20 and smaller), for a Container Apps workload profiles environment: the agent
  pipeline jobs of steps with `vpc: attach` (`container_apps_subnets` is the agent's default subnet) and other Container Apps in the same
  environment. One environment per subnet, delegated to `Microsoft.App/environments`, minimum /27, 14 IPs reserved
  (a /26 has 50 usable IPs, about 250 Consumption replicas; a /27 18, about 90; room for revision changes included), the
  size can't change once an environment uses it.
  The module creates the internal Container Apps environment (Consumption profile) in each Container Apps subnet, the agent
  finds it by subnet and waits until it is ready (creating a VNet environment takes 10+ minutes). Its infrastructure
  resource group is `<subnet name>-<location>` instead of an `ME_...` name, job logs go to the `<prefix>-container-apps` Log
  Analytics workspace.

#### Database subnets

  - Default mode: a database subnet is only created for /19 and larger VNets.
    - For /20 and smaller, private endpoints go to the private subnet or a custom subnet. Set database_subnets if you want a database subnet anyway.
  - Spoke mode: always has a database subnet.
  - Database subnets are never delegated. They hold private endpoints only.

#### How a /21 or /22 is split

  The VNet is split into two halves:

  - First half:
    - public /26
    - intra /26
    - AKS /27
    - Container Apps /27 (18 usable IPs)
    - AGC /24
    - the rest is spare
  - Second half: private

#### Minimum VNet size

  - Automatic split: needs /22 or larger. A /23 has no room for the AGC /24 alongside the other tiers.
  - Absolute minimum: vpc_cidr must be /24 or larger.
  - /23 and /24: set every subnet list explicitly, and use [] for tiers you don't need. Otherwise the automatic tiers fill the VNet and overlap.
    - In spoke mode this also means control_agc_subnets, service_agc_subnets, control_pe_subnets, service_pe_subnets and
      database_subnets; spoke mode needs room for two AGC /24s, so a /23 is too small for anything else and a /24 can't be
      spoke.
  - /24: only works without AGC (public_agc_subnets: []), because the AGC subnet alone takes a whole /24.
  - Fixed costs (about 320 addresses before any node): AGC /24, AKS /27 and Container Apps /27.


| vpc_cidr | public | intra | private | database | public-agc | aks | container apps | spare |
|---|---|---|---|---|---|---|---|---|
| /16 | /25 | /25 | /18 | /19 | /24 | /26 | /26 | /18 + /19 + /20 + /21 + /22 + /23 + /24 |
| /17 | /25 | /25 | /19 | /20 | /24 | /26 | /26 | /19 + /20 + /21 + /22 + /23 + /24 |
| /18 | /25 | /25 | /20 | /21 | /24 | /26 | /26 | /20 + /21 + /22 + /23 + /24 |
| /19 | /25 | /25 | /21 | /22 | /24 | /26 | /26 | /21 + /22 + /23 + /24 |
| /20 | /26 | /26 | /22 | none | /24 | /27 | /27 | /22 + /23 + /23 + /24 + /25 |
| /21 | /26 | /26 | /22 | none | /24 | /27 | /27 | /26 + /23 |
| /22 | /26 | /26 | /23 | none | /24 | /27 | /27 | /26 |

Spare, as the largest free blocks: for /20 and larger the fourth quarter (in a /20 also the unused database eighth) and
the rest of the first quarter after public and intra. The free room in the delegated block comes on top.

Spoke mode (`subnet_split_mode: spoke`, see below, /21 or larger): no intra, the node subnet named compute, always a
database subnet, a control-agc and a service-agc /24 (the first two /24s of the second eighth, in a /21 after the
public-agc /24) and a control-pe and a service-pe /28 (the first two /28s of the intra slot, in a /21 after aks and
Container Apps). Spare: the fourth quarter and the rest of the first quarter after
public, the private endpoint subnets, control-agc and service-agc:

| vpc_cidr | public | compute | database | public-agc | aks | container apps | control-agc | service-agc | control-pe | service-pe | spare |
|---|---|---|---|---|---|---|---|---|---|---|---|
| /16 | /25 | /18 | /19 | /24 | /26 | /26 | /24 | /24 | /28 | /28 | /18 + /20 + /20 + /21 + /21 + /22 + /22 + /23 + /23 + /24 + /26 + /27 |
| /17 | /25 | /19 | /20 | /24 | /26 | /26 | /24 | /24 | /28 | /28 | /19 + /21 + /21 + /22 + /22 + /23 + /23 + /24 + /26 + /27 |
| /18 | /25 | /20 | /21 | /24 | /26 | /26 | /24 | /24 | /28 | /28 | /20 + /22 + /22 + /23 + /23 + /24 + /26 + /27 |
| /19 | /25 | /21 | /22 | /24 | /26 | /26 | /24 | /24 | /28 | /28 | /21 + /23 + /23 + /24 + /26 + /27 |
| /20 | /26 | /22 | /23 | /24 | /27 | /27 | /24 | /24 | /28 | /28 | /22 + /24 + /25 + /27 |
| /21 | /26 | /22 | /26 | /24 | /27 | /27 | /24 | /24 | /28 | /28 | /27 |

Any subnet list can be set explicitly. The split does not shift for explicit lists, so overlapping subnets are
rejected at plan time. Extra Container Apps subnets (a separate Container Apps environment for a step) are extra entries
in `container_apps_subnets`, usually taken from the spare quarter; extra AGC subnets (more azure/agc modules) are custom
subnets (below).

### subnet_split_mode ###
`default` (the layout above): nodes, internal load balancer frontends (`type: LoadBalancer` Services) and the
internal gateway's private endpoint share the private subnet (the AGCs have their own delegated subnets: public-agc,
the internal gateway's custom /24), and the outputs `control_subnets`, `service_subnets` and `compute_subnets` all point
to it. `spoke` is for a VNet peered to a hub that
owns the firewall, gateways and DNS: the hub's rules are written per role range. It adds two AGC /24s in
the second eighth (after the public-agc /24 in a /21) and their two private endpoint /28s in the intra slot (in a /21
after aks and Container Apps), names the node subnet `<prefix>-compute-<n>`, creates no intra subnet (a hub spoke has no attachment
subnet, the peering attaches the VNet) and always creates a database subnet. The CIDRs of public, the delegated subnets
and (for /19 and larger) database are the same as in default mode. Switching an existing VNet to spoke drops the
intra subnet (in a /21 it must be empty: database takes its range), adds control-agc, service-agc, control-pe and
service-pe and
renames the node subnet, which Azure does by recreating it (new `private_subnets` id, so aks replaces the cluster): pin
the old name with `private_subnet_names` on a VNet with a running cluster. A /20 also gets the database subnet in a
former spare range.

| Subnet | Use | Output | /16 (10.156.0.0) | /20 (10.30.0.0) |
|---|---|---|---|---|
| control-agc-0 (/24) | AGC of the internal gateway (azure/agc, the platform's internal gateway) | `control_subnets`, `control_agc_subnet_names` (variables `control_agc_subnets`, `control_agc_subnet_names`) | 10.156.32.0/24 | 10.30.2.0/24 |
| service-agc-0 (/24) | AGC of the service gateway (azure/agc, other internal applications) | `service_subnets`, `service_agc_subnet_names` (variables `service_agc_subnets`, `service_agc_subnet_names`) | 10.156.33.0/24 | 10.30.3.0/24 |
| control-pe-0 (/28) | private endpoint of the internal gateway (the clients' IP) | `control_pe_subnets` | 10.156.0.128/28 | 10.30.0.64/28 |
| service-pe-0 (/28) | private endpoint of the service gateway | `service_pe_subnets` | 10.156.0.144/28 | 10.30.0.80/28 |
| compute-0 | nodes (the private subnet, named compute in spoke mode) | `compute_subnets`, `private_subnets` | unchanged | unchanged |
| intra | none in spoke mode (`intra_subnets` for one anyway) | `intra_subnets` = `[]` | | |

Both are delegated to `Microsoft.ServiceNetworking/trafficControllers` and hold nothing else: each is the association
subnet of one bring-your-own Application Gateway for Containers with a private frontend (azure/agc module, one AGC per
/24 with CNI Overlay), so the internal and the service gateway are separate load balancers in separate subnets like the
aws spoke ALBs. Their clients connect to the private endpoints in control-pe and service-pe (azure/agc
`private_endpoint_subnet_id`), so each gateway has its own entry range and its own load balancer range, apart from each
other and from compute; private endpoint network policies are on there, so NSGs and route tables apply to the private
endpoints. Internal Azure load balancers (`type: LoadBalancer` Services) stay in the compute subnet. None of them has a
NAT gateway or egress route. In default mode the outputs `control_pe_subnets` and `service_pe_subnets` are empty. More AGCs take more /24s in `custom_subnets` (delegation
`Microsoft.ServiceNetworking/trafficControllers`), from the free tail of the delegated block for /19 and larger (e.g.
10.156.162.0/24 in a /16) or from the spare quarter. In a /21
control-agc and service-agc take 10.156.2.0/24 and 10.156.3.0/24 after the public-agc /24, control-pe and service-pe
10.156.0.192/28 and 10.156.0.208/28, database a /26 in the intra slot 10.156.0.64/26 (59 private endpoints) and compute
stays /22. A /22 or smaller has no room for the two /24s: spoke mode
there needs explicit `control_agc_subnets`, `service_agc_subnets`, `private_subnets` and `database_subnets`, checked at plan
time. Spoke mode needs at least one control-agc and one service-agc subnet; these lists and the private endpoint lists
(`control_pe_subnets`, `service_pe_subnets`) can also be set explicitly in either mode (`*_subnet_names` as for the
other tiers) and the outputs then point to them. In default mode the internal
gateway's AGC /24 is a custom subnet (see below).

### Custom subnets ###
`custom_subnets` adds subnets for anything else, keyed by a short name (lowercase letters and digits, the agent reads
the outputs as `custom_subnets[key]`): `cidr` (usually from the spare quarter), optional `delegation` with `delegation_actions` (default the join
action), `name` (default `<prefix>-<key>`, for fixed names like `AzureBastionSubnet`), `network_security_group`
/ `route_table` for an empty NSG and route table associated with the subnet whose rules and routes the delegated
service manages, and `nat_gateway` for internet egress like the private subnets (the NAT gateway, or the egress route
with `egress_next_hop_ip`; not together with `route_table`, a subnet has one route table). Without `nat_gateway` a
custom subnet has no internet egress, which is what SQL Managed Instance, Bastion and private endpoint subnets need; a
Container Apps environment or VMs in a custom subnet need it. Azure SQL Managed Instance needs the delegation with the network policies actions, the NSG and the
route table, minimum /27 (a /24 fits 24 Business Critical instances in one VM group, 20 zone redundant), and no NAT gateway (not supported by
Managed Instance); Postgres, MySQL, Azure SQL Database and Redis use private endpoints in the database subnets instead.
A service that needs a delegated subnet (SQL Managed Instance, a VNet-integrated Postgres or MySQL flexible server) takes
a custom subnet, never a database subnet. An Application Gateway for Containers of an azure/agc module (the internal
gateway in default mode, further AGCs in both modes) takes a custom /24 delegated to
`Microsoft.ServiceNetworking/trafficControllers`, checked to be exactly /24 like the public-agc, control-agc and service-agc subnets.

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
            agcinternal:
              cidr: "10.156.162.0/24"
              delegation: Microsoft.ServiceNetworking/trafficControllers
```

Outputs `custom_subnets` (ids) and `custom_subnet_cidrs` are maps by key, e.g. `{{ .toutput.vpc.custom_subnets[mssql] }}`.

Default outbound access is off on all subnets: the private and Container Apps subnets egress through the NAT gateway
(`nat_static_ip_count` public IPs, about 64k SNAT ports each; `nat_gateway_sku` StandardV2 is zone redundant, Standard
single zone, changing it replaces the gateway and its IPs; `nat_idle_timeout_minutes`), intra and database subnets have
no internet access and public subnets only through a resource's own public IP or load balancer. With
`enable_nat_gateway: false` the private and Container Apps subnets have no internet egress at all: AKS
(`outbound_type: userAssignedNATGateway`) fails and pipeline jobs can't pull images. To egress through a firewall/NVA
(e.g. a peered hub) instead, set `enable_nat_gateway: false`, `egress_next_hop_ip` (0.0.0.0/0 route on the private and
Container Apps subnets) and aks `outbound_type: userDefinedRouting`.

Subnet names default to `<prefix>-<type>-<n>`, override them with `*_subnet_names` (Container Apps subnet names are
also the environment names: lowercase letters, digits, hyphens, max 60). Removing or reordering Container Apps subnets
recreates the environments after the removed one, a rare operation the agent handles by itself: it finds the
environment by subnet on the next run and waits for it. The private and
Container Apps subnets have service endpoints for Storage and Key Vault (`*_subnet_service_endpoints`): these services then
see the subnet's private IP instead of the NAT IP, so IP firewall rules need VNet rules instead. Container Apps environments
are zone redundant by default (`container_apps_zone_redundancy_enabled`, create time only).

Plan time checks: `vpc_cidr` an IPv4 CIDR of /24 or larger, `subnet_split_mode` default or spoke, overlapping subnets, subnets outside the VNet, AGC exactly
/24, unique subnet names across all tiers and custom subnets, `custom_subnets` keys and `nat_gateway` / `route_table`
exclusion and key format (1-40 lowercase letters or digits), `egress_next_hop_ip` not with `enable_nat_gateway`,
explicit public_agc, aks and Container Apps lists below /22, spoke mode below /21 with explicit lists and at least one
control-agc and one service-agc subnet, `unique_suffix` 8 lowercase letters or digits. Azure itself requires the aks subnet /28 or larger, Container Apps /27 or larger and Container Apps
subnet names that are valid Container Apps environment names (lowercase).

Outputs include `*_subnet_cidrs` (private, public, intra, database, public-agc, aks, Container Apps), `public_agc_subnets`, `control_pe_subnets` / `service_pe_subnets` (+ `_cidr_blocks`,
empty in default mode), `nat_static_ips`
and the names modules shared between clouds use (`name`, `vpc_cidr_block`,
`*_subnets_cidr_blocks`, `*_subnet_names`, `nat_public_ips`, `compute_subnets(_cidr_blocks)` = the private subnets,
and `control/service_subnets(_cidr_blocks)` = the control-agc and service-agc subnets when they exist (spoke mode or
explicit lists), otherwise the private subnets,
`elasticache_subnets_cidr_blocks` = database), e.g. for platform-apis. Outputs
of disabled resources (`nat_gateway_id`, `egress_route_table_id`, `vpc_flow_log_*`) are empty strings, not null, so
agent inputs referencing them always render.

VNet flow logs (`enable_flow_log`, default true): all IP flows of the VNet go to a dedicated storage account
`<prefix alnum>fl<agent uniqueSuffix>` (Entra ID only, Microsoft-managed keys, network access only for trusted Azure services and the
private subnets) with `flow_log_retention_days` retention (7), or to an existing storage account in the same
region given as `flow_log_storage_account_id` (e.g. a central log archive; the module then creates none and the account's
own network rules apply; `flow_log_retention_days` still sets the flow log retention). The flow log resource itself lives in the
region's Network Watcher resource group (`NetworkWatcherRG`/`NetworkWatcher_<location>`, one per region and subscription,
override with `network_watcher_name`/`network_watcher_resource_group_name`). `flow_log_traffic_analytics_enabled`
adds Traffic Analytics into a `<prefix>-flow-log` Log Analytics workspace (charged per GB).
Deleting the VNet (also by deleting its resource group) deletes its flow log in `NetworkWatcherRG` too, so a nuke leaves
nothing behind there.
- Why a dedicated account with Microsoft-managed keys: rotating a customer-managed key of the storage account stops
  the flow logs until they are disabled and enabled again ([VNet flow logs, Storage account: "Self-managed key
  rotation"](https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#storage-account)). The agent
  storage account rotates its key every 18 months and the kms keys rotate with `key_rotation_period`, so the logs
  would stop without any error. Customer-managed keys are set per storage account; encryption scopes per container are
  not documented to avoid this, so they aren't used here. Roadmap: an optional kms key without rotation only for the
  flow logs (compliance scanners flag keys without a rotation policy, a manual rotation still stops the logs).
- Not logged: Container Apps (the Container Apps environment), SQL Managed Instance and PostgreSQL/MySQL flexible server
  traffic ([incompatible services](https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#incompatible-services)).
- Cost: flow logs collected per GB, 5 GB/month free per subscription, plus storage
  ([pricing](https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#pricing)).

Fixed issue: azurerm 5.7.0 ended a Container Apps environment delete with "polling support for the Content-Type \"\"
was not implemented" although the environment was gone
([hashicorp/terraform-provider-azurerm#33433](https://github.com/hashicorp/terraform-provider-azurerm/issues/33433)),
fixed in azurerm 5.9.0 (go-azure-sdk update); the module pins a later version.

Not supported yet (roadmap): IPv6 / dual stack.

### Examples ###
```
    modules:
      - name: vpc
        source: azure/vpc
        inputs:
          vpc_cidr: "10.156.0.0/16"
```
Automatic subnets of `10.156.0.0/16` down to `10.156.0.0/20` (container-apps-0 also gets the Container Apps environment
`<prefix>-container-apps-0`, the spare range is for `custom_subnets` and explicit extra subnets):

| Subnet | /16 | /17 | /18 | /19 | /20 |
|---|---|---|---|---|---|
| public-0 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/26 |
| intra-0 | 10.156.0.128/25 | 10.156.0.128/25 | 10.156.0.128/25 | 10.156.0.128/25 | 10.156.0.64/26 |
| free after intra-0 | 10.156.1.0/24, 10.156.2.0/23, 10.156.4.0/22, 10.156.8.0/21, 10.156.16.0/20, 10.156.32.0/19 (10.156.1.0 to 10.156.63.255) | 10.156.1.0/24, 10.156.2.0/23, 10.156.4.0/22, 10.156.8.0/21, 10.156.16.0/20 (10.156.1.0 to 10.156.31.255) | 10.156.1.0/24, 10.156.2.0/23, 10.156.4.0/22, 10.156.8.0/21 (10.156.1.0 to 10.156.15.255) | 10.156.1.0/24, 10.156.2.0/23, 10.156.4.0/22 (10.156.1.0 to 10.156.7.255) | 10.156.0.128/25, 10.156.1.0/24, 10.156.2.0/23 (10.156.0.128 to 10.156.3.255) |
| private-0 | 10.156.64.0/18 | 10.156.32.0/19 | 10.156.16.0/20 | 10.156.8.0/21 | 10.156.4.0/22 |
| database-0 | 10.156.128.0/19 | 10.156.64.0/20 | 10.156.32.0/21 | 10.156.16.0/22 | none |
| public-agc-0 | 10.156.160.0/24 | 10.156.80.0/24 | 10.156.40.0/24 | 10.156.20.0/24 | 10.156.10.0/24 |
| aks-0 | 10.156.161.0/26 | 10.156.81.0/26 | 10.156.41.0/26 | 10.156.21.0/26 | 10.156.11.0/27 |
| container-apps-0 | 10.156.161.64/26 | 10.156.81.64/26 | 10.156.41.64/26 | 10.156.21.64/26 | 10.156.11.32/27 |
| free in the delegated block | 10.156.161.128/25, 10.156.162.0/23, 10.156.164.0/22, 10.156.168.0/21, 10.156.176.0/20 (10.156.161.128 to 10.156.191.255) | 10.156.81.128/25, 10.156.82.0/23, 10.156.84.0/22, 10.156.88.0/21 (10.156.81.128 to 10.156.95.255) | 10.156.41.128/25, 10.156.42.0/23, 10.156.44.0/22 (10.156.41.128 to 10.156.47.255) | 10.156.21.128/25, 10.156.22.0/23 (10.156.21.128 to 10.156.23.255) | 10.156.11.64/26, 10.156.11.128/25 (10.156.11.64 to 10.156.11.255) |
| spare | 10.156.192.0/18 | 10.156.96.0/19 | 10.156.48.0/20 | 10.156.24.0/21 | 10.156.8.0/23, 10.156.12.0/22 |

A /21 and a /22 use the small layout, here `10.156.0.0/21` and `10.156.0.0/22`:

| Subnet | /21 | /22 |
|---|---|---|
| public-0 | 10.156.0.0/26 | 10.156.0.0/26 |
| intra-0 | 10.156.0.64/26 | 10.156.0.64/26 |
| aks-0 | 10.156.0.128/27 | 10.156.0.128/27 |
| container-apps-0 | 10.156.0.160/27 | 10.156.0.160/27 |
| public-agc-0 | 10.156.1.0/24 | 10.156.1.0/24 |
| private-0 | 10.156.4.0/22 | 10.156.2.0/23 |
| spare | 10.156.0.192/26, 10.156.2.0/23 | 10.156.0.192/26 |

The aks /27 holds up to three clusters (a cluster reserves at least 9 of its 27 usable IPs), a SQL Managed Instance
takes a /27 or larger custom subnet from the spare range.

Spoke mode (`subnet_split_mode: spoke`), `10.156.0.0/16` down to `/20`: public, database and the delegated subnets keep
their default-mode addresses, the node subnet is named compute-0, control-pe and service-pe take the first two /28s of
the intra slot after public, control and service take the first two /24s of the second eighth and the rest of it stays
spare:

| Subnet | /16 | /17 | /18 | /19 | /20 |
|---|---|---|---|---|---|
| public-0 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/25 | 10.156.0.0/26 |
| control-pe-0 | 10.156.0.128/28 | 10.156.0.128/28 | 10.156.0.128/28 | 10.156.0.128/28 | 10.156.0.64/28 |
| service-pe-0 | 10.156.0.144/28 | 10.156.0.144/28 | 10.156.0.144/28 | 10.156.0.144/28 | 10.156.0.80/28 |
| free after service-pe-0 | 10.156.0.160/27, 10.156.0.192/26, 10.156.1.0/24, 10.156.2.0/23, 10.156.4.0/22, 10.156.8.0/21, 10.156.16.0/20 (10.156.0.160 to 10.156.31.255) | 10.156.0.160/27, 10.156.0.192/26, 10.156.1.0/24, 10.156.2.0/23, 10.156.4.0/22, 10.156.8.0/21 (10.156.0.160 to 10.156.15.255) | 10.156.0.160/27, 10.156.0.192/26, 10.156.1.0/24, 10.156.2.0/23, 10.156.4.0/22 (10.156.0.160 to 10.156.7.255) | 10.156.0.160/27, 10.156.0.192/26, 10.156.1.0/24, 10.156.2.0/23 (10.156.0.160 to 10.156.3.255) | 10.156.0.96/27, 10.156.0.128/25, 10.156.1.0/24 (10.156.0.96 to 10.156.1.255) |
| control-agc-0 | 10.156.32.0/24 | 10.156.16.0/24 | 10.156.8.0/24 | 10.156.4.0/24 | 10.156.2.0/24 |
| aks-0 | 10.156.161.0/26 | 10.156.81.0/26 | 10.156.41.0/26 | 10.156.21.0/26 | 10.156.11.0/27 |
| container-apps-0 | 10.156.161.64/26 | 10.156.81.64/26 | 10.156.41.64/26 | 10.156.21.64/26 | 10.156.11.32/27 |
| service-agc-0 | 10.156.33.0/24 | 10.156.17.0/24 | 10.156.9.0/24 | 10.156.5.0/24 | 10.156.3.0/24 |
| public-agc-0 | 10.156.160.0/24 | 10.156.80.0/24 | 10.156.40.0/24 | 10.156.20.0/24 | 10.156.10.0/24 |
| compute-0 | 10.156.64.0/18 | 10.156.32.0/19 | 10.156.16.0/20 | 10.156.8.0/21 | 10.156.4.0/22 |
| database-0 | 10.156.128.0/19 | 10.156.64.0/20 | 10.156.32.0/21 | 10.156.16.0/22 | 10.156.8.0/23 |
| free in the delegated block | 10.156.161.128/25, 10.156.162.0/23, 10.156.164.0/22, 10.156.168.0/21, 10.156.176.0/20 (10.156.161.128 to 10.156.191.255) | 10.156.81.128/25, 10.156.82.0/23, 10.156.84.0/22, 10.156.88.0/21 (10.156.81.128 to 10.156.95.255) | 10.156.41.128/25, 10.156.42.0/23, 10.156.44.0/22 (10.156.41.128 to 10.156.47.255) | 10.156.21.128/25, 10.156.22.0/23 (10.156.21.128 to 10.156.23.255) | 10.156.11.64/26, 10.156.11.128/25 (10.156.11.64 to 10.156.11.255) |
| spare | 10.156.34.0/23, 10.156.36.0/22, 10.156.40.0/21, 10.156.48.0/20 (10.156.34.0 to 10.156.63.255), 10.156.192.0/18 | 10.156.18.0/23, 10.156.20.0/22, 10.156.24.0/21 (10.156.18.0 to 10.156.31.255), 10.156.96.0/19 | 10.156.10.0/23, 10.156.12.0/22 (10.156.10.0 to 10.156.15.255), 10.156.48.0/20 | 10.156.6.0/23, 10.156.24.0/21 | 10.156.12.0/22 |

Spoke mode in a /21: control-agc and service-agc after the public-agc /24, database a /26 in the intra slot, control-pe and
service-pe in the spare /26 after aks and Container Apps, compute as in default mode (a /22 spoke needs explicit lists):

| Subnet | /21 |
|---|---|
| public-0 | 10.156.0.0/26 |
| database-0 | 10.156.0.64/26 |
| aks-0 | 10.156.0.128/27 |
| container-apps-0 | 10.156.0.160/27 |
| control-pe-0 | 10.156.0.192/28 |
| service-pe-0 | 10.156.0.208/28 |
| public-agc-0 | 10.156.1.0/24 |
| control-agc-0 | 10.156.2.0/24 |
| service-agc-0 | 10.156.3.0/24 |
| compute-0 | 10.156.4.0/22 |
| spare | 10.156.0.224/27 |

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
