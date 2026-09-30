locals {
  # First half private, second half split into 4 parts
  private_subnets  = var.private_subnets == null ? [cidrsubnet(var.vpc_cidr, 1, 0)] : var.private_subnets
  public_subnets   = var.public_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 0)] : var.public_subnets
  intra_subnets    = var.intra_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 1)] : var.intra_subnets
  database_subnets = var.database_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 2)] : var.database_subnets
  # Fourth part: AGC /24, API server /28, pipeline /27
  agc_block         = cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 3)
  agc_subnets       = var.agc_subnets == null ? [cidrsubnet(local.agc_block, 24 - tonumber(split("/", local.agc_block)[1]), 0)] : var.agc_subnets
  apiserver_subnets = var.apiserver_subnets == null ? [cidrsubnet(local.agc_block, 28 - tonumber(split("/", local.agc_block)[1]), 16)] : var.apiserver_subnets
  pipeline_subnets  = var.pipeline_subnets == null ? [cidrsubnet(local.agc_block, 27 - tonumber(split("/", local.agc_block)[1]), 9)] : var.pipeline_subnets

  # Overlaps are only rejected at apply by Azure, check them at plan time
  all_subnets = flatten([for type, cidrs in {
    private   = local.private_subnets
    public    = local.public_subnets
    intra     = local.intra_subnets
    database  = local.database_subnets
    agc       = local.agc_subnets
    apiserver = local.apiserver_subnets
    pipeline  = local.pipeline_subnets
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
