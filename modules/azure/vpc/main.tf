locals {
  vpc_bits = tonumber(split("/", var.vpc_cidr)[1])
  # Quarters: public and intra; private; database and the delegated block (AGC /24, aks /26, Container Apps /26); the last quarter is spare for custom_subnets
  # Public /25, aks and Container Apps /26 for /19 and larger, public /26, aks and Container Apps /27 for /20 to /22
  # /21 and /22: halves, public /26, intra /26, aks /27, Container Apps /27, AGC /24 and the rest spare in the first half, private the second half
  # No database subnet for /20 and smaller, smaller than /22: no room for the AGC /24, the delegated subnet lists must be explicit
  # Spoke split mode: the node subnet is named compute, no intra subnet, database also for /20 and /21, a control /24 (internal load balancers of the platform) and a service /24 (other internal load balancers) at the start of the intra slot; /21: control and service after the AGC /24, database /26 in the intra slot; smaller than /21: the lists must be explicit
  small           = local.vpc_bits > 20
  spoke           = var.subnet_split_mode == "spoke"
  delegated_block = cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 2), 1, 1)
  delegated_bits  = local.vpc_bits + 3

  public_subnets         = var.public_subnets == null ? [cidrsubnet(var.vpc_cidr, (local.vpc_bits > 19 ? 26 : 25) - local.vpc_bits, 0)] : var.public_subnets
  intra_subnets          = var.intra_subnets == null ? (local.spoke ? [] : [local.small ? cidrsubnet(var.vpc_cidr, 26 - local.vpc_bits, 1) : cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 0), 1, 1)]) : var.intra_subnets
  private_subnets        = var.private_subnets == null ? [local.small ? cidrsubnet(var.vpc_cidr, 1, 1) : cidrsubnet(var.vpc_cidr, 2, 1)] : var.private_subnets
  database_subnets       = var.database_subnets == null ? (local.spoke ? [local.small ? cidrsubnet(var.vpc_cidr, 26 - local.vpc_bits, 1) : cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 2), 1, 0)] : (local.vpc_bits > 19 ? [] : [cidrsubnet(cidrsubnet(var.vpc_cidr, 2, 2), 1, 0)])) : var.database_subnets
  public_agc_subnets     = var.public_agc_subnets == null ? try([local.small ? cidrsubnet(var.vpc_cidr, 24 - local.vpc_bits, 1) : cidrsubnet(local.delegated_block, 24 - local.delegated_bits, 0)], []) : var.public_agc_subnets
  aks_subnets            = var.aks_subnets == null ? try([local.small ? cidrsubnet(var.vpc_cidr, 27 - local.vpc_bits, 4) : cidrsubnet(cidrsubnet(local.delegated_block, 24 - local.delegated_bits, 1), local.vpc_bits > 19 ? 3 : 2, 0)], []) : var.aks_subnets
  container_apps_subnets = var.container_apps_subnets == null ? try([local.small ? cidrsubnet(var.vpc_cidr, 27 - local.vpc_bits, 5) : cidrsubnet(cidrsubnet(local.delegated_block, 24 - local.delegated_bits, 1), local.vpc_bits > 19 ? 3 : 2, 1)], []) : var.container_apps_subnets
  control_subnets        = var.control_subnets == null ? (local.spoke ? try([local.small ? cidrsubnet(var.vpc_cidr, 24 - local.vpc_bits, 2) : cidrsubnet(cidrsubnet(var.vpc_cidr, 3, 1), 21 - local.vpc_bits, 0)], []) : []) : var.control_subnets
  service_subnets        = var.service_subnets == null ? (local.spoke ? try([local.small ? cidrsubnet(var.vpc_cidr, 24 - local.vpc_bits, 3) : cidrsubnet(cidrsubnet(var.vpc_cidr, 3, 1), 21 - local.vpc_bits, 1)], []) : []) : var.service_subnets

  private_names        = [for i, cidr in local.private_subnets : try(var.private_subnet_names[i], "${var.prefix}-${local.spoke ? "compute" : "private"}-${i}")]
  public_names         = [for i, cidr in local.public_subnets : try(var.public_subnet_names[i], "${var.prefix}-public-${i}")]
  intra_names          = [for i, cidr in local.intra_subnets : try(var.intra_subnet_names[i], "${var.prefix}-intra-${i}")]
  database_names       = [for i, cidr in local.database_subnets : try(var.database_subnet_names[i], "${var.prefix}-database-${i}")]
  public_agc_names     = [for i, cidr in local.public_agc_subnets : try(var.public_agc_subnet_names[i], "${var.prefix}-public-agc-${i}")]
  aks_names            = [for i, cidr in local.aks_subnets : try(var.aks_subnet_names[i], "${var.prefix}-aks-${i}")]
  container_apps_names = [for i, cidr in local.container_apps_subnets : try(var.container_apps_subnet_names[i], "${var.prefix}-container-apps-${i}")]
  control_names        = [for i, cidr in local.control_subnets : try(var.control_subnet_names[i], "${var.prefix}-control-${i}")]
  service_names        = [for i, cidr in local.service_subnets : try(var.service_subnet_names[i], "${var.prefix}-service-${i}")]
  custom_names         = { for key, subnet in var.custom_subnets : key => coalesce(subnet.name, "${var.prefix}-${key}") }
  subnet_names         = concat(local.private_names, local.public_names, local.intra_names, local.database_names, local.public_agc_names, local.aks_names, local.container_apps_names, local.control_names, local.service_names, values(local.custom_names), local.pod_names)

  flow_log_storage_account_id = try(coalesce(var.flow_log_storage_account_id, one(azurerm_storage_account.flow_log[*].id)), "")

  # Azure CNI Pod Subnet: pod subnets in their own address space (pod_subnets_cidr) or in the first half of the spare quarter (enable_pod_subnets),
  # one subnet unless listed (the first is the aks default), equal split among the entries without cidr, explicit cidrs anywhere in the VNet
  pod_range        = var.pod_subnets_cidr != null ? var.pod_subnets_cidr : (var.enable_pod_subnets && !local.small ? cidrsubnet(var.vpc_cidr, 3, 6) : null)
  pod_subnet_specs = length(var.pod_subnets) > 0 ? var.pod_subnets : (local.pod_range == null ? [] : [{ cidr = null, dedicated_nat_gateway = false }])
  pod_subnet_split = [for i, subnet in local.pod_subnet_specs : i if subnet.cidr == null]
  pod_subnets      = compact([for i, subnet in local.pod_subnet_specs : subnet.cidr != null ? subnet.cidr : (local.pod_range != null ? cidrsubnet(local.pod_range, ceil(log(length(local.pod_subnet_split), 2)), index(local.pod_subnet_split, i)) : "")])
  pod_names        = [for i in range(length(local.pod_subnets)) : "${var.prefix}-pods-${i}"]
  pod_nat_subnets  = [for i, subnet in local.pod_subnet_specs : i if subnet.dedicated_nat_gateway]
  pod_nat_assoc    = var.enable_nat_gateway ? range(length(local.pod_subnets)) : local.pod_nat_subnets
  address_space    = compact([var.vpc_cidr, var.pod_subnets_cidr])

  # Plan time validations
  all_subnets = concat(flatten([for type, cidrs in {
    private        = local.private_subnets
    public         = local.public_subnets
    intra          = local.intra_subnets
    database       = local.database_subnets
    public_agc     = local.public_agc_subnets
    aks            = local.aks_subnets
    container_apps = local.container_apps_subnets
    control        = local.control_subnets
    service        = local.service_subnets
  } : [for i, cidr in cidrs : { name = "${type}-${i}", cidr = cidr }]]), [for key, subnet in var.custom_subnets : { name = "custom-${key}", cidr = subnet.cidr }], [for i, cidr in local.pod_subnets : { name = "pods-${i}", cidr = cidr }])
  subnet_ranges = [for s in concat([for cidr in local.address_space : { name = "vpc", cidr = cidr }], local.all_subnets) : merge(s, {
    start = sum([for i, octet in split(".", cidrhost(s.cidr, 0)) : tonumber(octet) * pow(256, 3 - i)])
    size  = pow(2, 32 - tonumber(split("/", s.cidr)[1]))
  })]
  space_ranges       = slice(local.subnet_ranges, 0, length(local.address_space))
  subnet_ranges_only = slice(local.subnet_ranges, length(local.address_space), length(local.subnet_ranges))
  subnets_outside_vpc = [for s in local.subnet_ranges_only : "${s.name} ${s.cidr}"
  if !anytrue([for r in local.space_ranges : s.start >= r.start && s.start + s.size <= r.start + r.size])]
  pod_space_overlaps_vpc = length(local.space_ranges) > 1 && local.space_ranges[0].start < local.space_ranges[1].start + local.space_ranges[1].size && local.space_ranges[1].start < local.space_ranges[0].start + local.space_ranges[0].size
  subnet_overlaps = flatten([for i, a in local.subnet_ranges_only : [
    for j, b in local.subnet_ranges_only : "${a.name} ${a.cidr} / ${b.name} ${b.cidr}"
    if j > i && a.start < b.start + b.size && b.start < a.start + a.size
  ]])

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })
}

resource "azurerm_virtual_network" "this" {
  name                = var.prefix
  location            = var.location
  resource_group_name = var.resource_group_name
  address_space       = local.address_space
  tags                = local.tags

  lifecycle {
    precondition {
      condition     = length(local.subnets_outside_vpc) == 0
      error_message = "Subnets outside the VNet address space ${join(", ", local.address_space)}: ${join(", ", local.subnets_outside_vpc)}."
    }
    precondition {
      condition     = local.pod_range != null || alltrue([for subnet in var.pod_subnets : subnet.cidr != null])
      error_message = "pod_subnets without a cidr need pod_subnets_cidr or enable_pod_subnets (/20 or larger) to split."
    }
    precondition {
      condition     = var.egress_next_hop_ip == null || !anytrue([for subnet in var.pod_subnets : subnet.dedicated_nat_gateway])
      error_message = "dedicated_nat_gateway and egress_next_hop_ip exclude each other: the 0.0.0.0/0 route to the appliance bypasses the NAT gateway."
    }
    precondition {
      condition     = var.egress_next_hop_ip == null || !var.enable_nat_gateway
      error_message = "egress_next_hop_ip needs enable_nat_gateway = false: the 0.0.0.0/0 route to the appliance bypasses the NAT gateway."
    }
    precondition {
      condition     = !(var.enable_pod_subnets && var.pod_subnets_cidr != null)
      error_message = "enable_pod_subnets and pod_subnets_cidr exclude each other: pod subnets come from the spare quarter or from the second address space."
    }
    precondition {
      condition     = !var.enable_pod_subnets || !local.small
      error_message = "enable_pod_subnets needs a /20 or larger vpc_cidr: set pod_subnets_cidr or pod_subnets with explicit cidrs."
    }
    precondition {
      condition     = !local.pod_space_overlaps_vpc
      error_message = "pod_subnets_cidr ${join(", ", slice(local.address_space, 1, length(local.address_space)))} overlaps vpc_cidr ${var.vpc_cidr}."
    }
    precondition {
      condition     = alltrue([for cidr in local.pod_subnets : tonumber(split("/", cidr)[1]) >= 16])
      error_message = "Pod subnets can be at most /16 (dynamic IP allocation): ${join(", ", local.pod_subnets)}."
    }
    precondition {
      condition     = local.vpc_bits <= 22 || (var.public_agc_subnets != null && var.aks_subnets != null && var.container_apps_subnets != null)
      error_message = "vpc_cidr is smaller than /22: set public_agc_subnets, aks_subnets and container_apps_subnets explicitly (or [])."
    }
    precondition {
      condition     = !local.spoke || local.vpc_bits <= 21 || (var.control_subnets != null && var.service_subnets != null && var.private_subnets != null && var.database_subnets != null)
      error_message = "subnet_split_mode spoke smaller than /21: set control_subnets, service_subnets, private_subnets and database_subnets explicitly (no room for the control and service /24s)."
    }
    precondition {
      condition     = !local.spoke || (length(local.control_subnets) > 0 && length(local.service_subnets) > 0)
      error_message = "subnet_split_mode spoke needs at least one control and one service subnet."
    }
    precondition {
      condition     = alltrue([for cidr in local.public_agc_subnets : tonumber(split("/", cidr)[1]) == 24])
      error_message = "Application Gateway for Containers subnets must be exactly /24: ${join(", ", local.public_agc_subnets)}."
    }
    precondition {
      condition     = length(distinct(local.subnet_names)) == length(local.subnet_names)
      error_message = "Subnet names must be unique: ${join(", ", [for n in distinct(local.subnet_names) : n if length([for m in local.subnet_names : m if m == n]) > 1])}."
    }
    precondition {
      condition     = alltrue([for key, subnet in var.custom_subnets : !(subnet.nat_gateway && subnet.route_table)])
      error_message = "custom_subnets nat_gateway and route_table exclude each other: the egress route table and an own route table can't both be associated."
    }
    precondition {
      condition     = length(local.subnet_overlaps) == 0
      error_message = "Overlapping subnets: ${join(", ", local.subnet_overlaps)}. The automatic split does not shift when a subnet list is set, so give the overlapping lists explicitly too."
    }
  }
}

resource "azurerm_subnet" "private" {
  count                = length(local.private_subnets)
  name                 = local.private_names[count.index]
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.this.name
  address_prefixes     = [local.private_subnets[count.index]]
  # No implicit internet egress, private subnets use the NAT gateway
  default_outbound_access_enabled = false

  dynamic "service_endpoint" {
    for_each = var.private_subnet_service_endpoints
    content {
      service = service_endpoint.value
    }
  }
}

resource "azurerm_subnet" "public" {
  count                           = length(local.public_subnets)
  name                            = local.public_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.public_subnets[count.index]]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "intra" {
  count                           = length(local.intra_subnets)
  name                            = local.intra_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.intra_subnets[count.index]]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "database" {
  count                           = length(local.database_subnets)
  name                            = local.database_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.database_subnets[count.index]]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "public_agc" {
  count                           = length(local.public_agc_subnets)
  name                            = local.public_agc_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.public_agc_subnets[count.index]]
  default_outbound_access_enabled = false

  delegation {
    name = "agc-delegation"

    service_delegation {
      name    = "Microsoft.ServiceNetworking/trafficControllers"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_subnet" "aks" {
  count                           = length(local.aks_subnets)
  name                            = local.aks_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.aks_subnets[count.index]]
  default_outbound_access_enabled = false

  delegation {
    name = "aks-apiserver-delegation"

    service_delegation {
      name    = "Microsoft.ContainerService/managedClusters"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_subnet" "container_apps" {
  count                           = length(local.container_apps_subnets)
  name                            = local.container_apps_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.container_apps_subnets[count.index]]
  default_outbound_access_enabled = false

  dynamic "service_endpoint" {
    for_each = var.container_apps_subnet_service_endpoints
    content {
      service = service_endpoint.value
    }
  }

  delegation {
    name = "container-apps-delegation"

    service_delegation {
      name    = "Microsoft.App/environments"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_log_analytics_workspace" "container_apps" {
  count               = length(local.container_apps_subnets) > 0 ? 1 : 0
  name                = "${var.prefix}-container-apps"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = local.tags
}

resource "azurerm_container_app_environment" "container_apps" {
  count                              = length(local.container_apps_subnets)
  name                               = azurerm_subnet.container_apps[count.index].name
  location                           = var.location
  resource_group_name                = var.resource_group_name
  infrastructure_subnet_id           = azurerm_subnet.container_apps[count.index].id
  infrastructure_resource_group_name = "${azurerm_subnet.container_apps[count.index].name}-${var.location}"
  internal_load_balancer_enabled     = true
  zone_redundancy_enabled            = var.container_apps_zone_redundancy_enabled
  logs_destination                   = "log-analytics"
  log_analytics_workspace_id         = azurerm_log_analytics_workspace.container_apps[0].id
  tags                               = local.tags

  workload_profile {
    name                  = "Consumption"
    workload_profile_type = "Consumption"
  }

  # azurerm 5.7.0 delete ends with a polling error although the delete succeeded, re-run the destroy:
  # https://github.com/hashicorp/terraform-provider-azurerm/issues/33433
  timeouts {
    delete = "60m"
  }

  depends_on = [
    azurerm_subnet_nat_gateway_association.container_apps,
    azurerm_nat_gateway_public_ip_association.this,
    azurerm_subnet_route_table_association.egress_container_apps,
  ]
}

# Load balancer frontends only: no NAT gateway or egress route, the backends answer from the node subnet
resource "azurerm_subnet" "control" {
  count                           = length(local.control_subnets)
  name                            = local.control_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.control_subnets[count.index]]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "service" {
  count                           = length(local.service_subnets)
  name                            = local.service_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.service_subnets[count.index]]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "custom" {
  for_each                        = var.custom_subnets
  name                            = local.custom_names[each.key]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [each.value.cidr]
  default_outbound_access_enabled = false

  dynamic "delegation" {
    for_each = each.value.delegation == null ? [] : [each.value.delegation]
    content {
      name = "${each.key}-delegation"

      service_delegation {
        name    = delegation.value
        actions = each.value.delegation_actions
      }
    }
  }
}

# Empty NSG and route table that the delegated service manages, e.g. SQL Managed Instance requires both
resource "azurerm_network_security_group" "custom" {
  for_each            = { for key, subnet in var.custom_subnets : key => subnet if subnet.network_security_group }
  name                = azurerm_subnet.custom[each.key].name
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags

  lifecycle {
    ignore_changes = [security_rule]
  }
}

resource "azurerm_subnet_network_security_group_association" "custom" {
  for_each                  = azurerm_network_security_group.custom
  subnet_id                 = azurerm_subnet.custom[each.key].id
  network_security_group_id = each.value.id
}

resource "azurerm_route_table" "custom" {
  for_each            = { for key, subnet in var.custom_subnets : key => subnet if subnet.route_table }
  name                = azurerm_subnet.custom[each.key].name
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags

  lifecycle {
    ignore_changes = [route]
  }
}

resource "azurerm_subnet_route_table_association" "custom" {
  for_each       = azurerm_route_table.custom
  subnet_id      = azurerm_subnet.custom[each.key].id
  route_table_id = each.value.id
}

# Azure CNI Pod Subnet: AKS delegates the pod subnets to itself, egress like the nodes through the shared NAT gateway, an
# own NAT gateway (dedicated_nat_gateway: own egress IP) or the egress route
resource "azurerm_subnet" "pods" {
  count                           = length(local.pod_subnets)
  name                            = local.pod_names[count.index]
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.pod_subnets[count.index]]
  default_outbound_access_enabled = false

  dynamic "service_endpoint" {
    for_each = var.private_subnet_service_endpoints
    content {
      service = service_endpoint.value
    }
  }

  delegation {
    name = "aks-pods-delegation"

    service_delegation {
      name    = "Microsoft.ContainerService/managedClusters"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_public_ip" "pods_nat" {
  count               = length(local.pod_nat_subnets)
  name                = "${var.prefix}-pods-${local.pod_nat_subnets[count.index]}-nat"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = var.nat_gateway_sku
  tags                = local.tags
}

resource "azurerm_nat_gateway" "pods" {
  count                   = length(local.pod_nat_subnets)
  name                    = "${var.prefix}-pods-${local.pod_nat_subnets[count.index]}"
  location                = var.location
  resource_group_name     = var.resource_group_name
  sku_name                = var.nat_gateway_sku
  idle_timeout_in_minutes = var.nat_idle_timeout_minutes
  tags                    = local.tags
}

resource "azurerm_nat_gateway_public_ip_association" "pods" {
  count                = length(local.pod_nat_subnets)
  nat_gateway_id       = azurerm_nat_gateway.pods[count.index].id
  public_ip_address_id = azurerm_public_ip.pods_nat[count.index].id
}

resource "azurerm_subnet_nat_gateway_association" "pods" {
  count          = length(local.pod_nat_assoc)
  subnet_id      = azurerm_subnet.pods[local.pod_nat_assoc[count.index]].id
  nat_gateway_id = contains(local.pod_nat_subnets, local.pod_nat_assoc[count.index]) ? azurerm_nat_gateway.pods[index(local.pod_nat_subnets, local.pod_nat_assoc[count.index])].id : azurerm_nat_gateway.this[0].id
}

resource "azurerm_subnet_route_table_association" "egress_pods" {
  count          = var.egress_next_hop_ip != null ? length(local.pod_subnets) : 0
  subnet_id      = azurerm_subnet.pods[count.index].id
  route_table_id = azurerm_route_table.egress[0].id
}

# Egress like the private subnets: the NAT gateway, or the egress route with egress_next_hop_ip
resource "azurerm_subnet_nat_gateway_association" "custom" {
  for_each       = { for key, subnet in var.custom_subnets : key => subnet if subnet.nat_gateway && var.enable_nat_gateway }
  subnet_id      = azurerm_subnet.custom[each.key].id
  nat_gateway_id = azurerm_nat_gateway.this[0].id
}

resource "azurerm_subnet_route_table_association" "egress_custom" {
  for_each       = { for key, subnet in var.custom_subnets : key => subnet if subnet.nat_gateway && var.egress_next_hop_ip != null }
  subnet_id      = azurerm_subnet.custom[each.key].id
  route_table_id = azurerm_route_table.egress[0].id
}

resource "azurerm_public_ip" "nat" {
  count               = var.enable_nat_gateway ? var.nat_static_ip_count : 0
  name                = count.index == 0 ? "${var.prefix}-nat" : "${var.prefix}-nat-${count.index}"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = var.nat_gateway_sku
  tags                = local.tags
}

resource "azurerm_nat_gateway" "this" {
  count                   = var.enable_nat_gateway ? 1 : 0
  name                    = var.prefix
  location                = var.location
  resource_group_name     = var.resource_group_name
  sku_name                = var.nat_gateway_sku
  idle_timeout_in_minutes = var.nat_idle_timeout_minutes
  tags                    = local.tags
}

resource "azurerm_nat_gateway_public_ip_association" "this" {
  count                = var.enable_nat_gateway ? var.nat_static_ip_count : 0
  nat_gateway_id       = azurerm_nat_gateway.this[0].id
  public_ip_address_id = azurerm_public_ip.nat[count.index].id
}

resource "azurerm_subnet_nat_gateway_association" "private" {
  count          = var.enable_nat_gateway ? length(local.private_subnets) : 0
  subnet_id      = azurerm_subnet.private[count.index].id
  nat_gateway_id = azurerm_nat_gateway.this[0].id
}

resource "azurerm_subnet_nat_gateway_association" "container_apps" {
  count          = var.enable_nat_gateway ? length(local.container_apps_subnets) : 0
  subnet_id      = azurerm_subnet.container_apps[count.index].id
  nat_gateway_id = azurerm_nat_gateway.this[0].id
}

# Egress through a firewall/NVA instead of the NAT gateway
resource "azurerm_route_table" "egress" {
  count               = var.egress_next_hop_ip != null ? 1 : 0
  name                = "${var.prefix}-egress"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags

  route {
    name                   = "default"
    address_prefix         = "0.0.0.0/0"
    next_hop_type          = "VirtualAppliance"
    next_hop_in_ip_address = var.egress_next_hop_ip
  }
}

resource "azurerm_subnet_route_table_association" "egress_private" {
  count          = var.egress_next_hop_ip != null ? length(local.private_subnets) : 0
  subnet_id      = azurerm_subnet.private[count.index].id
  route_table_id = azurerm_route_table.egress[0].id
}

resource "azurerm_subnet_route_table_association" "egress_container_apps" {
  count          = var.egress_next_hop_ip != null ? length(local.container_apps_subnets) : 0
  subnet_id      = azurerm_subnet.container_apps[count.index].id
  route_table_id = azurerm_route_table.egress[0].id
}

# VNet flow logs to a dedicated storage account with Microsoft-managed keys: rotating a customer-managed key of the
# storage account stops the flow logs until they are disabled and enabled again:
# https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview#storage-account
resource "azurerm_storage_account" "flow_log" {
  count                           = var.enable_flow_log && var.flow_log_storage_account_id == null ? 1 : 0
  name                            = "${substr(replace(lower(var.prefix), "/[^a-z0-9]/", ""), 0, 14)}fl${var.unique_suffix}"
  resource_group_name             = var.resource_group_name
  location                        = var.location
  account_kind                    = "StorageV2"
  account_tier                    = "Standard"
  account_replication_type        = "LRS"
  min_tls_version                 = "TLS1_2"
  shared_access_key_enabled       = false
  default_to_oauth_authentication = true
  allow_nested_items_to_be_public = false
  tags                            = local.tags

  network_rules {
    default_action             = "Deny"
    bypass                     = ["AzureServices", "Logging", "Metrics"]
    virtual_network_subnet_ids = azurerm_subnet.private[*].id
  }
  depends_on = [azurerm_subnet_nat_gateway_association.private, azurerm_subnet_route_table_association.egress_private]
}

resource "azurerm_log_analytics_workspace" "flow_log" {
  count               = var.enable_flow_log && var.flow_log_traffic_analytics_enabled ? 1 : 0
  name                = "${var.prefix}-flow-log"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = local.tags
}

resource "azurerm_network_watcher_flow_log" "this" {
  count                = var.enable_flow_log ? 1 : 0
  name                 = var.prefix
  network_watcher_name = var.network_watcher_name != "" ? var.network_watcher_name : "NetworkWatcher_${var.location}"
  resource_group_name  = var.network_watcher_resource_group_name
  location             = var.location
  target_resource_id   = azurerm_virtual_network.this.id
  storage_account_id   = local.flow_log_storage_account_id
  enabled              = true
  version              = 2
  tags                 = local.tags

  retention_policy {
    enabled = true
    days    = var.flow_log_retention_days
  }

  dynamic "traffic_analytics" {
    for_each = var.flow_log_traffic_analytics_enabled ? [1] : []
    content {
      enabled               = true
      workspace_id          = azurerm_log_analytics_workspace.flow_log[0].workspace_id
      workspace_region      = var.location
      workspace_resource_id = azurerm_log_analytics_workspace.flow_log[0].id
      interval_in_minutes   = 10
    }
  }
}
