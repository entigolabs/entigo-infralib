locals {
  public_domains_count  = length([for k, v in var.domains : k if !v.private])
  private_domains_count = length([for k, v in var.domains : k if v.private])

  domains = {
    for k, d in var.domains : k => merge(d, {
      vpc_ids = length(d.vpc_ids) > 0 ? d.vpc_ids : var.vpc_ids
      # Defaults like aws-v2/route53
      default_public = d.default_public != null ? d.default_public : (
        length(var.domains) == 1 ? true : (!d.private && local.public_domains_count == 1 ? true : null)
      )
      default_private = d.default_private != null ? d.default_private : (
        length(var.domains) == 1 ? true : (d.private && local.private_domains_count == 1 ? true : null)
      )
      # Public twin of a private zone for ACME DNS-01 challenges (cert-manager)
      needs_validation_zone = d.private && d.create_zone && d.create_validation
      # NS records in the parent zone, also in another resource group (e.g. a root zone shared by environments)
      parent_zone_name = d.parent_zone_id != "" ? provider::azurerm::parse_resource_id(d.parent_zone_id).resource_name : ""
      parent_zone_rg   = d.parent_zone_id != "" ? provider::azurerm::parse_resource_id(d.parent_zone_id).resource_group_name : ""
    })
  }

  default_public_keys  = [for k, v in local.domains : k if v.default_public == true]
  default_private_keys = [for k, v in local.domains : k if v.default_private == true]

  zone_ids = {
    for k, v in local.domains : k => (
      v.private ? (v.create_zone ? azurerm_private_dns_zone.this[k].id : data.azurerm_private_dns_zone.existing[k].id) :
      (v.create_zone ? azurerm_dns_zone.this[k].id : data.azurerm_dns_zone.existing[k].id)
    )
  }

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })
}

resource "azurerm_dns_zone" "this" {
  for_each            = { for k, v in local.domains : k => v if !v.private && v.create_zone }
  name                = each.value.domain_name
  resource_group_name = var.resource_group_name
  tags = merge(local.tags, {
    DefaultPublic  = each.value.default_public == true ? "true" : "false"
    DefaultPrivate = each.value.default_private == true ? "true" : "false"
  })
}

resource "azurerm_private_dns_zone" "this" {
  for_each            = { for k, v in local.domains : k => v if v.private && v.create_zone }
  name                = each.value.domain_name
  resource_group_name = var.resource_group_name
  tags = merge(local.tags, {
    DefaultPublic  = each.value.default_public == true ? "true" : "false"
    DefaultPrivate = each.value.default_private == true ? "true" : "false"
  })
}

data "azurerm_dns_zone" "existing" {
  for_each            = { for k, v in local.domains : k => v if !v.private && !v.create_zone }
  name                = each.value.domain_name
  resource_group_name = var.resource_group_name
}

data "azurerm_private_dns_zone" "existing" {
  for_each            = { for k, v in local.domains : k => v if v.private && !v.create_zone }
  name                = each.value.domain_name
  resource_group_name = var.resource_group_name
}

resource "azurerm_private_dns_zone_virtual_network_link" "this" {
  for_each = merge([
    for k, v in local.domains : { for i, id in v.vpc_ids : "${k}-${i}" => { key = k, vpc_id = id } }
    if v.private && v.create_zone
  ]...)
  name                 = "${var.prefix}-${each.key}"
  private_dns_zone_id  = azurerm_private_dns_zone.this[each.value.key].id
  virtual_network_id   = each.value.vpc_id
  registration_enabled = false
  tags                 = local.tags
}

resource "azurerm_dns_zone" "validation" {
  for_each            = { for k, v in local.domains : k => v if v.needs_validation_zone }
  name                = each.value.domain_name
  resource_group_name = var.resource_group_name
  tags                = merge(local.tags, { Purpose = "ACME-Validation" })
}

resource "azurerm_dns_ns_record" "delegation" {
  for_each            = { for k, v in local.domains : k => v if v.parent_zone_id != "" && !v.private && v.create_zone }
  name                = trimsuffix(each.value.domain_name, ".${each.value.parent_zone_name}")
  zone_name           = each.value.parent_zone_name
  resource_group_name = each.value.parent_zone_rg
  ttl                 = 300
  records             = azurerm_dns_zone.this[each.key].name_servers
  tags                = local.tags

  lifecycle {
    precondition {
      condition     = endswith(lower(each.value.domain_name), ".${lower(each.value.parent_zone_name)}")
      error_message = "${each.value.domain_name} is not a subdomain of its parent zone ${each.value.parent_zone_name}."
    }
  }
}

resource "azurerm_dns_ns_record" "validation_delegation" {
  for_each            = { for k, v in local.domains : k => v if v.parent_zone_id != "" && v.needs_validation_zone }
  name                = trimsuffix(each.value.domain_name, ".${each.value.parent_zone_name}")
  zone_name           = each.value.parent_zone_name
  resource_group_name = each.value.parent_zone_rg
  ttl                 = 300
  records             = azurerm_dns_zone.validation[each.key].name_servers
  tags                = local.tags

  lifecycle {
    precondition {
      condition     = endswith(lower(each.value.domain_name), ".${lower(each.value.parent_zone_name)}")
      error_message = "${each.value.domain_name} is not a subdomain of its parent zone ${each.value.parent_zone_name}."
    }
  }
}
