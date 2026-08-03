output "cluster_id" {
  value = oci_containerengine_cluster.this.id
}

output "cluster_name" {
  value = oci_containerengine_cluster.this.name
}

output "kubernetes_version" {
  value = oci_containerengine_cluster.this.kubernetes_version
}

output "public_endpoint" {
  value = oci_containerengine_cluster.this.endpoints[0].public_endpoint
}

output "private_endpoint" {
  value = oci_containerengine_cluster.this.endpoints[0].private_endpoint
}

output "kubernetes_endpoint" {
  value = oci_containerengine_cluster.this.endpoints[0].kubernetes
}

output "main_node_pool_id" {
  value = try(module.main[0].node_pool_id, "")
}

output "mon_node_pool_id" {
  value = try(module.mon[0].node_pool_id, "")
}

output "tools_node_pool_id" {
  value = try(module.tools[0].node_pool_id, "")
}

output "lb_nsg_id" {
  value = oci_core_network_security_group.lb.id
}

output "main_min_size" {
  value = var.oke_main_min_size
}

output "main_max_size" {
  value = var.oke_main_max_size
}

output "mon_min_size" {
  value = var.oke_mon_min_size
}

output "mon_max_size" {
  value = var.oke_mon_max_size
}

output "tools_min_size" {
  value = var.oke_tools_min_size
}

output "tools_max_size" {
  value = var.oke_tools_max_size
}
