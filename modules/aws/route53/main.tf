locals {
  domain = var.parent_zone_id != "" ? data.aws_route53_zone.parent_zone[0].name : var.parent_domain
  pub_domain = var.create_public ? "${local.hname}.${local.domain}" : local.domain
  int_domain =var.create_private ? "${local.hname}-int.${local.domain}" : local.pub_domain

}

resource "aws_route53_zone" "pub" {
  count = var.create_public ? 1 : 0
  name = local.pub_domain
}

resource "aws_acm_certificate" "pub" {
  count = var.create_cert ? 1 : 0
  domain_name       = "*.${local.pub_domain}"
  validation_method = "DNS"
}

resource "aws_route53_record" "pub-cert" {
  for_each = var.create_cert ? {
    for dvo in aws_acm_certificate.pub[0].domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = local.pub_zone
}

resource "aws_acm_certificate_validation" "pub" {
  count = var.create_cert ? 1 : 0
  certificate_arn         = aws_acm_certificate.pub[0].arn
}

resource "aws_route53_zone" "int" {
  count = var.create_private ? 1 : 0
  name = local.int_domain
  vpc {
    vpc_id = data.aws_ssm_parameter.vpc_id.value
  }
  lifecycle {
    ignore_changes = [vpc]
  }
}

resource "aws_route53_zone" "int-cert" {
  count = var.create_cert && var.create_private ? 1 : 0
  name = local.int_domain
}

resource "aws_acm_certificate" "int" {
  count = var.create_cert && var.create_private ? 1 : 0
  domain_name       = "*.${local.int_domain}"
  validation_method = "DNS"
}

resource "aws_route53_record" "int-cert" {
  for_each = var.create_cert && var.create_private ? {
    for dvo in aws_acm_certificate.int[0].domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = aws_route53_zone.int-cert[0].zone_id
}

resource "aws_acm_certificate_validation" "int" {
  count = var.create_cert && var.create_private ? 1 : 0
  certificate_arn         = aws_acm_certificate.int[0].arn
}


resource "aws_route53_record" "pub-ns" {
  count = var.parent_zone_id != "" ? 1 : 0
  zone_id = data.aws_route53_zone.parent_zone[0].zone_id
  name = local.pub_domain
  type    = "NS"
  ttl     = "120"
  records = aws_route53_zone.pub[0].name_servers
}

resource "aws_route53_record" "int-cert-ns" {
  count = var.parent_zone_id != "" && var.create_private ? 1 : 0
  zone_id = data.aws_route53_zone.parent_zone[0].zone_id
  name = local.int_domain
  type    = "NS"
  ttl     = "120"
  records = aws_route53_zone.int-cert[0].name_servers
}

locals {
  zone = var.parent_zone_id != "" ? data.aws_route53_zone.parent_zone[0].zone_id : ""
  pub_zone = var.create_public ? aws_route53_zone.pub[0].zone_id : local.zone
  int_zone = var.create_private ? aws_route53_zone.int[0].zone_id : local.pub_zone
}

resource "aws_ssm_parameter" "pub" {
  count = var.create_public ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/route53/pub_zone_id"
  type  = "String"
  value = local.pub_zone
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "pub-domain" {
  count = var.create_public ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/route53/pub_domain"
  type  = "String"
  value = local.pub_domain
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "int" {
  name  = "/entigo-infralib/${local.hname}/route53/int_zone_id"
  type  = "String"
  value = local.int_zone
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "int-domain" {
  name  = "/entigo-infralib/${local.hname}/route53/int_domain"
  type  = "String"
  value = local.int_domain
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}
