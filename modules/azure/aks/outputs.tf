output "cluster_id" {
  value = module.aks.resource_id
}

output "cluster_name" {
  value = module.aks.name
}

output "cluster_version" {
  value = module.aks.current_kubernetes_version
}

output "cluster_endpoint" {
  value = "https://${coalesce(module.aks.fqdn, module.aks.private_fqdn)}:443"
}

output "cluster_certificate_authority_data" {
  value = module.aks.cluster_ca_certificate
}

output "oidc_issuer_url" {
  value = module.aks.oidc_issuer_profile_issuer_url
}

output "region" {
  value = var.location
}

output "resource_group_name" {
  value = var.resource_group_name
}

output "node_resource_group" {
  value = module.aks.node_resource_group_name
}

output "node_resource_group_id" {
  value = module.aks.node_resource_group_id
}

output "identity_principal_id" {
  value = azurerm_user_assigned_identity.aks.principal_id
}

output "kubelet_identity_object_id" {
  value = module.aks.kubelet_identity.objectId
}

output "kubelet_identity_client_id" {
  value = module.aks.kubelet_identity.clientId
}

output "kubernetes_version" {
  value = var.kubernetes_version
}

output "vnet_subnet_id" {
  value = var.vnet_subnet_id
}

output "os_sku" {
  value = var.os_sku
}

output "pod_cidr" {
  value = var.pod_cidr
}

output "service_cidr" {
  value = var.service_cidr
}

output "disk_encryption_set_id" {
  value = local.disk_encryption ? azurerm_disk_encryption_set.this[0].id : ""
}

output "alb_controller_client_id" {
  value = length(var.agc_subnet_ids) > 0 ? azurerm_user_assigned_identity.alb_controller[0].client_id : ""
}
