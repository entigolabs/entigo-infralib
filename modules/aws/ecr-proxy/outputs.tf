output "hub_registry" {
  description = "Registry URL for docker hub"
  value       = var.hub_username != "" && var.hub_token != "" ? "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_pull_through_cache_rule.hub[0].ecr_repository_prefix}" : ""
}

output "ghcr_registry" {
  description = "Registry URL for docker hub"
  value       = var.ghcr_username != "" && var.ghcr_token != "" ? "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_pull_through_cache_rule.ghcr[0].ecr_repository_prefix}" : ""
}

output "gcr_registry" {
  description = "Registry URL for docker hub"
  value       =  var.gcr_username != "" && var.gcr_token != "" ? "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_pull_through_cache_rule.gcr[0].ecr_repository_prefix}" : ""
}

output "k8s_registry" {
  description = "Registry URL for docker hub"
  value       = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_pull_through_cache_rule.k8s.ecr_repository_prefix}"
}

output "ecr_registry" {
  description = "Registry URL for docker hub"
  value       = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_pull_through_cache_rule.ecr.ecr_repository_prefix}"
}

output "quay_registry" {
  description = "Registry URL for docker hub"
  value       = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/${aws_ecr_pull_through_cache_rule.quay.ecr_repository_prefix}"
}
