# Get all shared resolver rules
data "aws_route53_resolver_rules" "shared" {  
  share_status = var.share_status
}

data "aws_route53_resolver_rule" "details" {
  for_each         = toset(data.aws_route53_resolver_rules.shared.resolver_rule_ids)
  resolver_rule_id = each.value
}
