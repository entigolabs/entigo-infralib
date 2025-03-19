# Filter out resolver rules that match excluded domains
locals {
  filtered_rules = {
    for rule in data.aws_route53_resolver_rule.details :
    rule.domain_name => rule.resolver_rule_id
    if !contains(var.exclude_domains, rule.domain_name)
  }
}

# Associate only valid shared resolver rules with the VPC
resource "aws_route53_resolver_rule_association" "shared_rules" {
  for_each         = local.filtered_rules
  resolver_rule_id = each.value
  vpc_id           = var.vpc_id
  name             = substr(replace("${var.prefix}-${each.key}", ".", "-"), 0, 63)
}
