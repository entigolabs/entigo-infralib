output "prefix" {
  value = var.prefix
}

output "node_pool_name" {
  value = module.aks_node_pool.name
}

output "node_pool_id" {
  value = module.aks_node_pool.resource_id
}

output "cluster_id" {
  value = var.cluster_id
}
