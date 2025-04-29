output "service_account_email" {
  value = google_service_account.crossplane.email
}

output "kubernetes_namespace" {
  value = var.kubernetes_namespace
}

output "kubernetes_service_account" {
  value = var.kubernetes_service_account
}

output "project_id" {
  value = data.google_client_config.this.project
}
