output "cluster_name" {
  value = module.eks.cluster_name
}

output "cluster_oidc_issuer_url" {
  value = module.eks.cluster_oidc_issuer_url
}

output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "cluster_certificate_authority_data" {
  value = module.eks.cluster_certificate_authority_data
}

output "cluster_version" {
  value = module.eks.cluster_version
}

output "oidc_provider" {
  value = module.eks.oidc_provider
}

output "oidc_provider_arn" {
  value = module.eks.oidc_provider_arn
}

output "eks_desired_size_map" {
  value = local.eks_desired_size_map
}

output "eks_min_size_map" {
  value = local.eks_min_size_map
}

output "eks_min_and_desired_size_map" {
  value = local.eks_min_and_desired_size_map
}

output "extra_desired_sizes" {
  value = local.extra_desired_sizes
}

output "extra_min_sizes" {
  value = local.extra_min_sizes
}

output "temp_map_1" {
  value = local.temp_map_1
}

output "temp_map_2" {
  value = local.temp_map_2
}
