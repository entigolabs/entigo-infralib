locals {
  # Count public and private domains
  public_domains_count  = length([for k, v in var.domains : k if !v.private])
  private_domains_count = length([for k, v in var.domains : k if v.private])
  # Determine effective VPC ID for each domain
  domains_with_defaults = {
    for key, domain in var.domains : key => {
      # Merge the original domain with our computed values
      domain_name                = domain.domain_name
      parent_zone_id             = domain.parent_zone_id
      create_zone                = domain.create_zone
      create_certificate         = domain.create_certificate
      certificate_authority_arn  = domain.certificate_authority_arn
      private                    = domain.private
      default_public             = domain.default_public
      default_private            = domain.default_private
      
      # Use domain-specific VPC ID if provided, otherwise use default module VPC ID
      effective_vpc_id           = domain.vpc_id != "" ? domain.vpc_id : var.vpc_id
      
      # Calculate default_public:
      # - Use user value if provided
      # - If only one domain total, set to true
      # - If only one public domain, set to true for that domain
      # - Otherwise keep as null (user must specify)
      default_public = domain.default_public != null ? domain.default_public : (
        length(var.domains) == 1 ? true : (
          !domain.private && local.public_domains_count == 1 ? true : null
        )
      )
      
      # Calculate default_private:
      # - Use user value if provided
      # - If only one domain total, set to true
      # - If only one private domain, set to true for that domain
      # - Otherwise keep as null (user must specify)
      default_private = domain.default_private != null ? domain.default_private : (
        length(var.domains) == 1 ? true : (
          domain.private && local.private_domains_count == 1 ? true : null
        )
      )
     # Flag to indicate if we need a validation zone for this domain
     needs_validation_zone = domain.private && domain.create_certificate && domain.create_zone && domain.certificate_authority_arn == ""
    }
  }
  
  # Validation: Check exactly one domain is default_public and one is default_private
  # (if any domains exist in each category)
  validation_rules = {
    one_default_public = (
      length([for k, v in local.domains_with_defaults : k if v.default_public == true]) == 1
    )
    one_default_private = (
      length([for k, v in local.domains_with_defaults : k if v.default_private == true]) == 1
    )
  }
  
  
  # Find the keys for default public and private zones
  default_public_key = [
    for k, v in local.domains_with_defaults : k 
    if v.default_public == true
  ][0]
  
  default_private_key = [
    for k, v in local.domains_with_defaults : k 
    if v.default_private == true
  ][0]
  
  effective_zone_ids = {
    for k, v in var.domains : k => (
      v.create_zone == false ? 
      data.aws_route53_zone.existing[k].zone_id : 
      aws_route53_zone.this[k].zone_id
    )
  }
}

resource "aws_route53_zone" "this" {
  for_each = {
    for k, v in var.domains : k => v
    if v.create_zone
  }
  
  name    = each.value.domain_name
  comment = "${var.prefix} - Managed by Terraform"
  dynamic "vpc" {
    for_each = each.value.private ? [1] : []
    content {
      vpc_id = local.domains_with_defaults[each.key].effective_vpc_id
    }
  }
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    DefaultPublic  = local.domains_with_defaults[each.key].default_public == true ? "true" : "false"
    DefaultPrivate = local.domains_with_defaults[each.key].default_private == true ? "true" : "false"
  }
  lifecycle {
    ignore_changes = [vpc]
    precondition {
      condition     = local.validation_rules.one_default_public && local.validation_rules.one_default_private
      error_message = "Validation failed: Make sure exactly one domain is set as default_public=true and exactly one is set as default_private=true"
    }
  }
}

#We assume the domains exist already that have "create_zone = false"
data "aws_route53_zone" "existing" {
  for_each = {
    for k, v in var.domains : k => v
    if v.create_zone == false
  }
  
  name         = each.value.domain_name
  private_zone = each.value.private
}

# Create NS records in parent zone for public zones when parent_zone_id is provided
resource "aws_route53_record" "ns_delegation" {
  for_each = {
    for k, v in var.domains : k => v
    if v.parent_zone_id != "" && !v.private && v.create_zone  # Only for public zones
  }
  
  zone_id = each.value.parent_zone_id
  name    = each.value.domain_name
  type    = "NS"
  ttl     = 300
  records = aws_route53_zone.this[each.key].name_servers
}


# Additional public validation zones for private domains that need certificate validation
resource "aws_route53_zone" "validation" {
  for_each = {
    for k, v in local.domains_with_defaults : k => v
    if v.needs_validation_zone
  }
  
  name    = each.value.domain_name
  comment = "${var.prefix} - ACM validation zone - Managed by Terraform"
  
  # No VPC association since this is deliberately public for ACM validation
  
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Purpose   = "ACM-Validation"
  }
}

# Create NS records in parent zone for the validation zones when needed
resource "aws_route53_record" "validation_ns_delegation" {
  for_each = {
    for k, v in local.domains_with_defaults : k => v
    if v.needs_validation_zone && v.parent_zone_id != ""
  }
  
  zone_id = each.value.parent_zone_id
  name    = each.value.domain_name
  type    = "NS"
  ttl     = 300
  records = aws_route53_zone.validation[each.key].name_servers
}

resource "aws_acm_certificate" "this" {
  for_each = {
    for k, v in var.domains : k => v
    if v.create_certificate
  }
  
  domain_name               = "*.${each.value.domain_name}"
  subject_alternative_names = [each.value.domain_name]
  
  # Conditionally set validation_method or certificate_authority_arn
  validation_method         = each.value.certificate_authority_arn == "" ? "DNS" : null
  certificate_authority_arn = each.value.certificate_authority_arn != "" ? each.value.certificate_authority_arn : null
  
  # Use the key algorithm from the domain configuration
  key_algorithm             = each.value.certificate_key_algorithm
  
  lifecycle {
    create_before_destroy = true
  }
  
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

# Create DNS validation records - using the appropriate zone based on domain type
resource "aws_route53_record" "validation" {
  for_each = {
    for dvo in flatten([
      for domain_key, cert in aws_acm_certificate.this : [
        for dvo in cert.domain_validation_options : {
          domain_key = domain_key
          domain_name = dvo.domain_name
          dvo = dvo
          is_private = local.domains_with_defaults[domain_key].private
          needs_validation_zone = local.domains_with_defaults[domain_key].needs_validation_zone
          create_zone = local.domains_with_defaults[domain_key].create_zone
        }
      ]
      # Only include certificates that have validation options (public certificates)
      if length(cert.domain_validation_options) > 0
    ]) : "${dvo.domain_key}_${dvo.domain_name}" => dvo
  }
  
  # Use the validation zone for private domains that need certificate validation,
  # otherwise use the primary zone
  zone_id = each.value.needs_validation_zone ? aws_route53_zone.validation[each.value.domain_key].zone_id : ( each.value.create_zone == false ? data.aws_route53_zone.existing[each.value.domain_key].zone_id :     aws_route53_zone.this[each.value.domain_key].zone_id  )
  name    = each.value.dvo.resource_record_name
  type    = each.value.dvo.resource_record_type
  records = [each.value.dvo.resource_record_value]
  ttl     = 60
  allow_overwrite = true
}

# Validate the public certificates
resource "aws_acm_certificate_validation" "this" {
  for_each = {
    for k, cert in aws_acm_certificate.this : k => cert
    if cert.validation_method == "DNS"
  }
  certificate_arn         = each.value.arn
}

