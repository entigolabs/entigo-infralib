output "google_project_iam_member" {
    value = resource.google_project_iam_member.external_dns
}

output "dns_zone" {
    value = trimsuffix(module.dns.domain, ".")
}
