output "prefix" {
  value = var.prefix
}

output "node_pool_name" {
  value = azurerm_kubernetes_cluster_node_pool.this.name
}

output "node_pool_id" {
  value = azurerm_kubernetes_cluster_node_pool.this.id
}

output "cluster_id" {
  value = var.cluster_id
}
