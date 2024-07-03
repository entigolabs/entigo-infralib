output "service_account_email" {
  value = google_service_account.crossplane.email
}

output "project_id" {
  value = data.google_client_config.this.project
}
