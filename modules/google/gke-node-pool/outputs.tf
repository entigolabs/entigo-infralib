output "prefix" {
  value = var.prefix
}

output "node_pool_name" {
  value = local.node_pool_name
}

output "node_pool_id" {
  value = module.gke_node_pool.id
}

output "cluster_name" {
  value = var.cluster_name
}

output "cluster_region" {
  value = var.cluster_region
}
