output "dns_zone" {
    value = trimsuffix(module.dns.domain, ".")
}
