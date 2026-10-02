locals {
  # First half private, second half split into 4 parts
  private_subnets  = var.private_subnets == null ? [cidrsubnet(var.vpc_cidr, 1, 0)] : var.private_subnets
  public_subnets   = var.public_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 0)] : var.public_subnets
  intra_subnets    = var.intra_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 1)] : var.intra_subnets
  database_subnets = var.database_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 2)] : var.database_subnets
  # Fourth part, delegated subnets
  delegated_block = cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 3)
  delegated_bits  = tonumber(split("/", local.delegated_block)[1])
  apiserver_bits  = max(24, local.delegated_bits + 5)
  # AGC /24, API server and pipeline next to each other: /24 each for a /16, /28 and /27 for a /20
  # Smaller than /20: no room, the delegated subnet lists must be explicit
  delegated_subnets = local.delegated_bits <= 23 ? cidrsubnets(local.delegated_block, 24 - local.delegated_bits, local.apiserver_bits - local.delegated_bits, min(27, local.apiserver_bits) - local.delegated_bits) : []
  agc_subnets       = var.agc_subnets == null ? try([local.delegated_subnets[0]], []) : var.agc_subnets
  apiserver_subnets = var.apiserver_subnets == null ? try([local.delegated_subnets[1]], []) : var.apiserver_subnets
  pipeline_subnets  = var.pipeline_subnets == null ? try([local.delegated_subnets[2]], []) : var.pipeline_subnets
  # Last quarter of the delegated block for MSSQL Managed Instances
  mssql_subnets = var.enable_mssql_subnets ? (var.mssql_subnets == null ? [cidrsubnet(local.delegated_block, 2, 3)] : var.mssql_subnets) : []

  # Plan time validations
  all_subnets = flatten([for type, cidrs in {
    private   = local.private_subnets
    public    = local.public_subnets
    intra     = local.intra_subnets
    database  = local.database_subnets
    agc       = local.agc_subnets
    apiserver = local.apiserver_subnets
    pipeline  = local.pipeline_subnets
    mssql     = local.mssql_subnets
  } : [for i, cidr in cidrs : { name = "${type}-${i}", cidr = cidr }]])
  subnet_ranges = [for s in concat([{ name = "vpc", cidr = var.vpc_cidr }], local.all_subnets) : merge(s, {
    start = sum([for i, octet in split(".", cidrhost(s.cidr, 0)) : tonumber(octet) * pow(256, 3 - i)])
    size  = pow(2, 32 - tonumber(split("/", s.cidr)[1]))
  })]
  vpc_range = local.subnet_ranges[0]
  subnets_outside_vpc = [for s in slice(local.subnet_ranges, 1, length(local.subnet_ranges)) : "${s.name} ${s.cidr}"
  if s.start < local.vpc_range.start || s.start + s.size > local.vpc_range.start + local.vpc_range.size]
  subnet_overlaps = flatten([for i, a in slice(local.subnet_ranges, 1, length(local.subnet_ranges)) : [
    for j, b in slice(local.subnet_ranges, 1, length(local.subnet_ranges)) : "${a.name} ${a.cidr} / ${b.name} ${b.cidr}"
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
  address_space       = [var.vpc_cidr]
  tags                = local.tags

  lifecycle {
    precondition {
      condition     = length(local.subnets_outside_vpc) == 0
      error_message = "Subnets outside vpc_cidr ${var.vpc_cidr}: ${join(", ", local.subnets_outside_vpc)}."
    }
    precondition {
      condition     = local.delegated_bits <= 23 || (var.agc_subnets != null && var.apiserver_subnets != null && var.pipeline_subnets != null)
      error_message = "vpc_cidr is smaller than /20: set agc_subnets, apiserver_subnets and pipeline_subnets explicitly (or [])."
    }
    precondition {
      condition     = alltrue([for cidr in local.mssql_subnets : tonumber(split("/", cidr)[1]) <= 27])
      error_message = "SQL Managed Instance subnets must be /27 or larger: ${join(", ", local.mssql_subnets)}."
    }
    precondition {
      condition     = length(local.subnet_overlaps) == 0
      error_message = "Overlapping subnets: ${join(", ", local.subnet_overlaps)}. The automatic split does not shift when a subnet list is set, so give the overlapping lists explicitly too."
    }
  }
}

resource "azurerm_subnet" "private" {
  count                = length(local.private_subnets)
  name                 = try(var.private_subnet_names[count.index], "${var.prefix}-private-${count.index}")
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
  name                            = try(var.public_subnet_names[count.index], "${var.prefix}-public-${count.index}")
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.public_subnets[count.index]]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "intra" {
  count                           = length(local.intra_subnets)
  name                            = try(var.intra_subnet_names[count.index], "${var.prefix}-intra-${count.index}")
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.intra_subnets[count.index]]
  default_outbound_access_enabled = false
}

resource "azurerm_subnet" "database" {
  count                           = length(local.database_subnets)
  name                            = try(var.database_subnet_names[count.index], "${var.prefix}-database-${count.index}")
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.database_subnets[count.index]]
  default_outbound_access_enabled = false

  dynamic "delegation" {
    for_each = var.database_subnet_delegation == null ? [] : [var.database_subnet_delegation]
    content {
      name = "database-delegation"

      service_delegation {
        name    = delegation.value
        actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
      }
    }
  }
}

resource "azurerm_subnet" "agc" {
  count                           = length(local.agc_subnets)
  name                            = try(var.agc_subnet_names[count.index], "${var.prefix}-agc-${count.index}")
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.agc_subnets[count.index]]
  default_outbound_access_enabled = false

  delegation {
    name = "agc-delegation"

    service_delegation {
      name    = "Microsoft.ServiceNetworking/trafficControllers"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_subnet" "apiserver" {
  count                           = length(local.apiserver_subnets)
  name                            = try(var.apiserver_subnet_names[count.index], "${var.prefix}-apiserver-${count.index}")
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.apiserver_subnets[count.index]]
  default_outbound_access_enabled = false

  delegation {
    name = "aks-apiserver-delegation"

    service_delegation {
      name    = "Microsoft.ContainerService/managedClusters"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_subnet" "pipeline" {
  count                           = length(local.pipeline_subnets)
  name                            = try(var.pipeline_subnet_names[count.index], "${var.prefix}-pipeline-${count.index}")
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.pipeline_subnets[count.index]]
  default_outbound_access_enabled = false

  delegation {
    name = "pipeline-delegation"

    service_delegation {
      name    = "Microsoft.App/environments"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_log_analytics_workspace" "pipeline" {
  count               = length(local.pipeline_subnets) > 0 ? 1 : 0
  name                = "${var.prefix}-pipeline"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = local.tags
}

resource "azurerm_container_app_environment" "pipeline" {
  count                              = length(local.pipeline_subnets)
  name                               = azurerm_subnet.pipeline[count.index].name
  location                           = var.location
  resource_group_name                = var.resource_group_name
  infrastructure_subnet_id           = azurerm_subnet.pipeline[count.index].id
  infrastructure_resource_group_name = "${azurerm_subnet.pipeline[count.index].name}-${var.location}"
  internal_load_balancer_enabled     = true
  logs_destination                   = "log-analytics"
  log_analytics_workspace_id         = azurerm_log_analytics_workspace.pipeline[0].id
  tags                               = local.tags

  workload_profile {
    name                  = "Consumption"
    workload_profile_type = "Consumption"
  }

  depends_on = [azurerm_subnet_nat_gateway_association.pipeline]
}

resource "azurerm_subnet" "mssql" {
  count                           = length(local.mssql_subnets)
  name                            = try(var.mssql_subnet_names[count.index], "${var.prefix}-mssql-${count.index}")
  resource_group_name             = var.resource_group_name
  virtual_network_name            = azurerm_virtual_network.this.name
  address_prefixes                = [local.mssql_subnets[count.index]]
  default_outbound_access_enabled = false

  delegation {
    name = "mssql-delegation"

    service_delegation {
      name = "Microsoft.Sql/managedInstances"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
        "Microsoft.Network/virtualNetworks/subnets/prepareNetworkPolicies/action",
        "Microsoft.Network/virtualNetworks/subnets/unprepareNetworkPolicies/action",
      ]
    }
  }
}

resource "azurerm_network_security_group" "mssql" {
  count               = length(local.mssql_subnets)
  name                = azurerm_subnet.mssql[count.index].name
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags

  lifecycle {
    ignore_changes = [security_rule]
  }
}

resource "azurerm_subnet_network_security_group_association" "mssql" {
  count                     = length(local.mssql_subnets)
  subnet_id                 = azurerm_subnet.mssql[count.index].id
  network_security_group_id = azurerm_network_security_group.mssql[count.index].id
}

resource "azurerm_route_table" "mssql" {
  count               = length(local.mssql_subnets)
  name                = azurerm_subnet.mssql[count.index].name
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags

  lifecycle {
    ignore_changes = [route]
  }
}

resource "azurerm_subnet_route_table_association" "mssql" {
  count          = length(local.mssql_subnets)
  subnet_id      = azurerm_subnet.mssql[count.index].id
  route_table_id = azurerm_route_table.mssql[count.index].id
}

resource "azurerm_public_ip" "nat" {
  count               = var.enable_nat_gateway ? var.nat_static_ip_count : 0
  name                = count.index == 0 ? "${var.prefix}-nat" : "${var.prefix}-nat-${count.index}"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = local.tags
}

resource "azurerm_nat_gateway" "this" {
  count               = var.enable_nat_gateway ? 1 : 0
  name                = var.prefix
  location            = var.location
  resource_group_name = var.resource_group_name
  sku_name            = "Standard"
  tags                = local.tags
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

resource "azurerm_subnet_nat_gateway_association" "pipeline" {
  count          = var.enable_nat_gateway ? length(local.pipeline_subnets) : 0
  subnet_id      = azurerm_subnet.pipeline[count.index].id
  nat_gateway_id = azurerm_nat_gateway.this[0].id
}
