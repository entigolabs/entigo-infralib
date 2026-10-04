## Opinionated module for Azure VNet creation ##

Creates a VNet with private, public, intra and database subnets, a delegated subnet for Application Gateway for
Containers (AGC), a delegated subnet for AKS API Server VNet Integration, a delegated pipeline subnet for Container
Apps (the agent jobs), a delegated subnet for Azure SQL Managed Instance and a NAT gateway for the private and pipeline
subnets.

### Automatic subnet calculations ###
If only `vpc_cidr` is given, the subnets are calculated from it:

* private: the first half of the VNet
* public, intra, database: the first three quarters of the second half
* the last quarter of the second half is the delegated block for the service-delegated subnets below, with free room
  for more of them (e.g. extra AGC /24s)
* agc: the first /24 of the delegated block (AGC with Azure CNI Overlay needs exactly a /24)
* apiserver: right after the AGC /24, a /24 for a /16 VNet (about 27 clusters at 9+ IPs each, room for blue/green
  cluster replacement), down to /28 for a /20
* pipeline: after the API server subnet, a /24 for a /16 VNet down to /27 for a /20, for a Container Apps workload
  profiles environment: the agent jobs of steps with `vpc: attach` (`pipeline_subnets` is the agent's default subnet)
  and other Container Apps in the same environment. One environment per subnet, delegated to
  `Microsoft.App/environments`, minimum /27 = 18 usable IPs (up to 90 Consumption replicas), the size can't change once
  an environment uses it.
  The module creates the internal Container Apps environment (Consumption profile) in each pipeline subnet, the agent
  finds it by subnet and waits until it is ready (creating a VNet environment takes 10+ minutes). Its infrastructure
  resource group is `<subnet name>-<location>` instead of an `ME_...` name, job logs go to the `<prefix>-pipeline` Log
  Analytics workspace.
* mssql: only with `enable_mssql_subnets: true`, the last quarter of the delegated block for Azure SQL Managed
  Instance (a /21 fits about 100 Business Critical instances)

The mssql subnets are delegated to `Microsoft.Sql/managedInstances`, each with its own NSG and route table in which the
service manages its rules and routes, and no NAT gateway (not supported by Managed Instance). Minimum /27, the size
can't change once used. Postgres, MySQL, Azure SQL Database and Redis use private endpoints in the database subnets
instead.

The minimum VNet size for the automatic split is /20. For smaller VNets set `agc_subnets`, `apiserver_subnets` and
`pipeline_subnets` explicitly (or `[]` without AKS or Container Apps); the other subnet lists usually need to be set too,
because the default tiers fill the rest of the VNet.

| vpc_cidr | private | public, intra, database | agc | apiserver | pipeline | mssql |
|---|---|---|---|---|---|---|
| /16 | /17 | /19 each | /24 | /24 | /24 | /21 |
| /18 | /19 | /21 each | /24 | /26 | /26 | /23 |
| /20 | /21 | /23 each | /24 | /28 | /27 | /25 |

Any subnet list can be set explicitly. The split does not shift for explicit lists, so overlapping subnets are
rejected at plan time.

Default outbound access is off on all subnets: the private and pipeline subnets egress through the NAT gateway
(`nat_static_ip_count` public IPs, about 64k SNAT ports each; `nat_gateway_sku` StandardV2 is zone redundant, Standard
single zone, changing it replaces the gateway and its IPs; `nat_idle_timeout_minutes`), intra and database subnets have
no internet access and public subnets only through a resource's own public IP or load balancer. With
`enable_nat_gateway: false` the private and pipeline subnets have no internet egress at all: AKS
(`outbound_type: userAssignedNATGateway`) fails and pipeline jobs can't pull images. To egress through a firewall/NVA
(e.g. a peered hub) instead, set `enable_nat_gateway: false`, `egress_next_hop_ip` (0.0.0.0/0 route on the private and
pipeline subnets) and aks `outbound_type: userDefinedRouting`.

Subnet names default to `<prefix>-<type>-<n>`, override them with `*_subnet_names`. Pipeline subnets and their
environments are keyed by name: with explicit `pipeline_subnet_names` (valid environment names: lowercase letters,
digits, hyphens, max 60) removing or reordering one doesn't recreate the others (10+ min each). The private and
pipeline subnets have service endpoints for Storage and Key Vault (`*_subnet_service_endpoints`): these services then
see the subnet's private IP instead of the NAT IP, so IP firewall rules need VNet rules instead. Pipeline environments
are zone redundant by default (`pipeline_zone_redundancy_enabled`, create time only).

Plan time checks: overlapping subnets, subnets outside the VNet, AGC exactly /24, unique pipeline subnet names,
`mssql_subnets` only with `enable_mssql_subnets`. Azure itself requires the API server subnet /28 or larger, pipeline
and mssql /27 or larger and pipeline subnet names that are valid Container Apps environment names (lowercase).

Outputs use both the google style (`*_subnet_cidrs`, `nat_static_ips`) and the aws/vpc names (`name`,
`*_subnets_cidr_blocks`, `*_subnet_names`, `nat_public_ips`, and `control/service/compute_subnets(_cidr_blocks)` =
private, `elasticache_subnets_cidr_blocks` = database) for modules shared between clouds, e.g. platform-apis.

VNet flow logs (`enable_flow_log`, default true like aws/vpc): all IP flows of the VNet go to a dedicated storage account
`<prefix alnum>fl<random>` (Entra ID only, Microsoft-managed keys, network access only for trusted Azure services and the
private subnets) with `flow_log_retention_days` retention (7, like aws). The flow log resource itself lives in the
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
  not documented to avoid this, so they aren't used here.
- Not logged: Container Apps (pipeline environment), SQL Managed Instance and PostgreSQL/MySQL flexible server
  traffic ([incompatible services](https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#incompatible-services)).
- Cost: flow logs collected per GB, 5 GB/month free per subscription, plus storage
  ([pricing](https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#pricing)).

Known issue: azurerm 5.7.0 ends a Container Apps environment delete with "polling support for the Content-Type \"\"
was not implemented" although the environment is gone; run the destroy again
([hashicorp/terraform-provider-azurerm#33433](https://github.com/hashicorp/terraform-provider-azurerm/issues/33433)).

Not supported yet (roadmap): IPv6 / dual stack (aws `enable_ipv6`) and a spoke split mode (aws `subnet_split_mode`).

### Example code ###
```
    modules:
      - name: vpc
        source: azure/vpc
        inputs:
          vpc_cidr: "10.156.0.0/16"
```
Will result in networks:

private-0 ( 10.156.0.0/17 )

public-0 ( 10.156.128.0/19 )

intra-0 ( 10.156.160.0/19 )

database-0 ( 10.156.192.0/19 )

agc-0 ( 10.156.224.0/24 )

apiserver-0 ( 10.156.225.0/24 )

pipeline-0 ( 10.156.226.0/24, Container Apps environment `<prefix>-pipeline-0` )

mssql-0 ( 10.156.248.0/21 ) with `enable_mssql_subnets: true`
