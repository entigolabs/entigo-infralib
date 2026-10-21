locals {
  subnet = provider::azurerm::parse_resource_id(var.subnet_id)

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })
}

data "azurerm_subnet" "this" {
  name                 = local.subnet.resource_name
  virtual_network_name = local.subnet.parent_resources["virtualNetworks"]
  resource_group_name  = local.subnet.resource_group_name
}

resource "azurerm_application_load_balancer" "this" {
  name                = var.prefix
  resource_group_name = var.resource_group_name
  location            = var.location
  tags                = local.tags
}

resource "azurerm_application_load_balancer_subnet_association" "this" {
  name                         = local.subnet.resource_name
  application_load_balancer_id = azurerm_application_load_balancer.this.id
  subnet_id                    = var.subnet_id
  tags                         = local.tags
}

# azurerm has no private frontend (publicNetworkAccess, association); the GA API 2026-03-01 is newer than azapi's
# embedded schemas (up to 2025-03-01-preview)
resource "azapi_resource" "frontend" {
  type                      = "Microsoft.ServiceNetworking/trafficControllers/frontends@2026-03-01"
  name                      = "private"
  parent_id                 = azurerm_application_load_balancer.this.id
  location                  = var.location
  tags                      = local.tags
  schema_validation_enabled = false
  body = {
    properties = {
      publicNetworkAccess = "Disabled"
      association = {
        id = azurerm_application_load_balancer_subnet_association.this.id
      }
    }
  }
  response_export_values = { fqdn = "properties.fqdn" }
}

# ALB controller: configures the AGC, the overlay extension config of CNI Overlay needs the subnet
resource "azurerm_role_assignment" "config_manager" {
  scope                            = azurerm_application_load_balancer.this.id
  role_definition_name             = "AppGw for Containers Configuration Manager"
  principal_id                     = var.alb_controller_principal_id
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "subnet" {
  scope                            = var.subnet_id
  role_definition_name             = "Network Contributor"
  principal_id                     = var.alb_controller_principal_id
  skip_service_principal_aad_check = true
}

# The private frontend is a Private Link resource: clients reach it through a private endpoint (group = frontend name)
resource "azurerm_private_endpoint" "this" {
  name                          = var.prefix
  resource_group_name           = var.resource_group_name
  location                      = var.location
  subnet_id                     = var.private_endpoint_subnet_id
  custom_network_interface_name = "${var.prefix}-nic"
  tags                          = local.tags

  private_service_connection {
    name                           = azapi_resource.frontend.name
    private_connection_resource_id = azurerm_application_load_balancer.this.id
    subresource_names              = [azapi_resource.frontend.name]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "alb"
    private_dns_zone_ids = [var.private_dns_zone_id]
  }
}
