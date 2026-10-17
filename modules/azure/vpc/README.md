## Opinionated module for Azure VNet creation ##

Creates a VNet with private, public, intra and database subnets, a delegated subnet for Application Gateway for
Containers (AGC), a delegated subnet for AKS API Server VNet Integration, a delegated Container Apps subnet (the agent
pipeline jobs), optional custom subnets (e.g. Azure SQL Managed Instance) and a NAT gateway for the private and
Container Apps subnets.

### Automatic subnet calculations ###
If only `vpc_cidr` is given, the subnets are calculated from it in quarters:

* first quarter: public and intra (an eighth each)
* second quarter: private
* third quarter: database (an eighth) and the delegated block (an eighth) for the service-delegated subnets below,
  with free room after them
* fourth quarter: spare, never used by the automatic split, for `custom_subnets` and explicit extra AGC or Container Apps
  subnets
* agc: the first /24 of the delegated block (AGC with Azure CNI Overlay needs exactly a /24, one AGC per subnet, up to
  5 frontends per AGC)
* apiserver: the /26 after the AGC /24 (59 usable IPs; a cluster reserves at least 9, so room for blue/green cluster
  replacement; more clusters: set `apiserver_subnets`)
* container-apps: the /26 after the API server subnet, for a Container Apps workload profiles environment: the agent
  pipeline jobs of steps with `vpc: attach` (`container_apps_subnets` is the agent's default subnet) and other Container Apps in the same
  environment. One environment per subnet, delegated to `Microsoft.App/environments`, minimum /27, 14 IPs reserved,
  about one IP per 10 Consumption replicas (a /26 is about 250 replicas), the size can't change once an environment
  uses it.
  The module creates the internal Container Apps environment (Consumption profile) in each Container Apps subnet, the agent
  finds it by subnet and waits until it is ready (creating a VNet environment takes 10+ minutes). Its infrastructure
  resource group is `<subnet name>-<location>` instead of an `ME_...` name, job logs go to the `<prefix>-container-apps` Log
  Analytics workspace.

A database subnet is only created for /19 and larger; for /20 and smaller private endpoints go to the private or a
custom subnet (set `database_subnets` for one anyway). A /21 or /22 is split in halves: the first half holds public
/26, intra /26, the API server /28 (one AKS cluster), the Container Apps /27 (18 usable IPs), the AGC /24 and the rest as
spare; the second half is private.

The minimum VNet size for the automatic split is /22: a /23 has no room for the AGC /24 next to the other tiers, and
`vpc_cidr` must be /24 or larger. For /23 and /24 set every subnet list explicitly (`[]` for tiers not needed): the
automatic tiers would fill the VNet and overlap. A /24 only works without AGC (`agc_subnets: []`), the AGC subnet alone
is a /24; the fixed costs are AGC /24, API server /28 and Container Apps /27, about 300 addresses before any node.

| vpc_cidr | public, intra | private | database | agc | apiserver | container apps | spare |
|---|---|---|---|---|---|---|---|
| /16 | /19 each | /18 | /19 | /24 | /26 | /26 | /18 |
| /17 | /20 each | /19 | /20 | /24 | /26 | /26 | /19 |
| /18 | /21 each | /20 | /21 | /24 | /26 | /26 | /20 |
| /19 | /22 each | /21 | /22 | /24 | /26 | /26 | /21 |
| /20 | /23 each | /22 | none | /24 | /26 | /26 | /23 + /22 |
| /21 | /26 each | /22 | none | /24 | /28 | /27 | /28 + /26 + /23 |
| /22 | /26 each | /23 | none | /24 | /28 | /27 | /28 + /26 |

Any subnet list can be set explicitly. The split does not shift for explicit lists, so overlapping subnets are
rejected at plan time. Extra AGC subnets (a second azure-gateway with its own AGC) and extra Container Apps subnets
(a separate Container Apps environment for a step) are extra entries in `agc_subnets` and `container_apps_subnets`,
usually taken from the spare quarter.

### Custom subnets ###
`custom_subnets` adds subnets for anything else, keyed by a short name (lowercase letters and digits, the agent reads
the outputs as `custom_subnets[key]`): `cidr`, optional `delegation` with `delegation_actions` (default the join
action), `name` (default `<prefix>-<key>`, for fixed names like `AzureBastionSubnet`), `network_security_group`
/ `route_table` for an empty NSG and route table associated with the subnet whose rules and routes the delegated
service manages, and `nat_gateway` for internet egress like the private subnets (the NAT gateway, or the egress route
with `egress_next_hop_ip`; not together with `route_table`, a subnet has one route table). Without `nat_gateway` a
custom subnet has no internet egress, which is what SQL Managed Instance, Bastion and private endpoint subnets need; a
Container Apps environment or VMs in a custom subnet need it. Azure SQL Managed Instance needs the delegation with the network policies actions, the NSG and the
route table, minimum /27 (a /24 fits about 20 Business Critical instances), and no NAT gateway (not supported by
Managed Instance); Postgres, MySQL, Azure SQL Database and Redis use private endpoints in the database subnets instead.

```
          custom_subnets:
            mssql:
              cidr: "10.156.192.0/24"
              delegation: Microsoft.Sql/managedInstances
              delegation_actions:
                - Microsoft.Network/virtualNetworks/subnets/join/action
                - Microsoft.Network/virtualNetworks/subnets/prepareNetworkPolicies/action
                - Microsoft.Network/virtualNetworks/subnets/unprepareNetworkPolicies/action
              network_security_group: true
              route_table: true
            bastion:
              name: AzureBastionSubnet
              cidr: "10.156.193.0/26"
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

Plan time checks: `vpc_cidr` an IPv4 CIDR of /24 or larger, overlapping subnets, subnets outside the VNet, AGC exactly
/24, unique subnet names across all tiers and custom subnets, `custom_subnets` keys and `nat_gateway` / `route_table`
exclusion. Azure itself requires the API server subnet /28 or larger, Container Apps /27 or larger and Container Apps
subnet names that are valid Container Apps environment names (lowercase).

Outputs use both the google style (`*_subnet_cidrs`, `nat_static_ips`) and the aws/vpc names (`name`, `vpc_cidr_block`,
`*_subnets_cidr_blocks`, `*_subnet_names`, `nat_public_ips`, and `control/service/compute_subnets(_cidr_blocks)` =
private, `elasticache_subnets_cidr_blocks` = database) for modules shared between clouds, e.g. platform-apis. Outputs
of disabled resources (`nat_gateway_id`, `egress_route_table_id`, `vpc_flow_log_*`) are empty strings, not null, so
agent inputs referencing them always render.

VNet flow logs (`enable_flow_log`, default true like aws/vpc): all IP flows of the VNet go to a dedicated storage account
`<prefix alnum>fl<agent uniqueSuffix>` (Entra ID only, Microsoft-managed keys, network access only for trusted Azure services and the
private subnets) with `flow_log_retention_days` retention (7, like aws), or to an existing storage account in the same
region given as `flow_log_storage_account_id` (e.g. a central log archive; the module then creates none and the account's
own network rules and retention apply). The flow log resource itself lives in the
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

Known issue: azurerm 5.7.0 ends a Container Apps environment delete with "polling support for the Content-Type \"\"
was not implemented" although the environment is gone; run the destroy again
([hashicorp/terraform-provider-azurerm#33433](https://github.com/hashicorp/terraform-provider-azurerm/issues/33433)).

Not supported yet (roadmap): IPv6 / dual stack (aws `enable_ipv6`) and a spoke split mode (aws `subnet_split_mode`).

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
| public-0 | 10.156.0.0/19 | 10.156.0.0/20 | 10.156.0.0/21 | 10.156.0.0/22 | 10.156.0.0/23 |
| intra-0 | 10.156.32.0/19 | 10.156.16.0/20 | 10.156.8.0/21 | 10.156.4.0/22 | 10.156.2.0/23 |
| private-0 | 10.156.64.0/18 | 10.156.32.0/19 | 10.156.16.0/20 | 10.156.8.0/21 | 10.156.4.0/22 |
| database-0 | 10.156.128.0/19 | 10.156.64.0/20 | 10.156.32.0/21 | 10.156.16.0/22 | none |
| agc-0 | 10.156.160.0/24 | 10.156.80.0/24 | 10.156.40.0/24 | 10.156.20.0/24 | 10.156.10.0/24 |
| apiserver-0 | 10.156.161.0/26 | 10.156.81.0/26 | 10.156.41.0/26 | 10.156.21.0/26 | 10.156.11.0/26 |
| container-apps-0 | 10.156.161.64/26 | 10.156.81.64/26 | 10.156.41.64/26 | 10.156.21.64/26 | 10.156.11.64/26 |
| free in the delegated block | 10.156.161.128/25 to 10.156.191.255 | 10.156.81.128/25 to 10.156.95.255 | 10.156.41.128/25 to 10.156.47.255 | 10.156.21.128/25 to 10.156.23.255 | 10.156.11.128/25 |
| spare | 10.156.192.0/18 | 10.156.96.0/19 | 10.156.48.0/20 | 10.156.24.0/21 | 10.156.8.0/23, 10.156.12.0/22 |

A /21 and a /22 use the small layout, here `10.156.0.0/21` and `10.156.0.0/22`:

| Subnet | /21 | /22 |
|---|---|---|
| public-0 | 10.156.0.0/26 | 10.156.0.0/26 |
| intra-0 | 10.156.0.64/26 | 10.156.0.64/26 |
| apiserver-0 | 10.156.0.128/28 | 10.156.0.128/28 |
| container-apps-0 | 10.156.0.160/27 | 10.156.0.160/27 |
| agc-0 | 10.156.1.0/24 | 10.156.1.0/24 |
| private-0 | 10.156.4.0/22 | 10.156.2.0/23 |
| spare | 10.156.0.144/28, 10.156.0.192/26, 10.156.2.0/23 | 10.156.0.144/28, 10.156.0.192/26 |

A second AKS cluster in a /21 or /22 takes the spare /28 next to the first as a second `apiserver_subnets` entry, a SQL
Managed Instance a /27 or larger custom subnet from the rest of the spare range.

A /23 or /24 needs every list explicit, e.g. `10.156.0.0/23` without Container Apps:
```
    modules:
      - name: vpc
        source: azure/vpc
        inputs:
          vpc_cidr: "10.156.0.0/23"
          public_subnets: '["10.156.0.0/27"]'
          intra_subnets: '["10.156.0.32/27"]'
          apiserver_subnets: '["10.156.0.64/28"]'
          container_apps_subnets: '[]'
          private_subnets: '["10.156.0.128/25"]'
          database_subnets: '[]'
          agc_subnets: '["10.156.1.0/24"]'
```
