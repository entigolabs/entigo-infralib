output "prefix" {
  value = var.prefix
}

output "service_agent_emails" {
  value = merge(
    { for service, identity in google_project_service_identity.this : service => identity.email },
    local.predefined_service_agent_emails
  )
}
