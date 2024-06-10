output "cluster_id" {
  description = "GKE cluster ID"
  value       = module.gke.cluster_id
}

output "cluster_endpoint" {
  description = "Endpoint for GKE control plane"
  value       = nonsensitive(module.gke.endpoint)
}

output "cluster_name" {
  description = "Google Kubernetes Cluster Name"
  value       = module.gke.name
}

output "region" {
  description = "GKE region"
  value       = module.gke.region
}
