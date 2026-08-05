variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the DNS zone."
  type        = string
}

variable "parent_domain" {
  description = "The domain this zone is a subdomain of, e.g. tarmo.entigo.dev. The parent zone lives elsewhere (e.g. Route53) - this module only creates the OCI-side zone and outputs its nameservers for manual NS delegation."
  type        = string
}

variable "subdomain_name" {
  description = "Subdomain label to create under parent_domain. Defaults to prefix if unset."
  type        = string
  default     = ""
}

# The wildcard certificate every app under this zone is served from. Supplied rather than
# created here, and that is not for want of trying: modules/k8s/oci-native-ingress-controller
# terminates TLS on an OCI load balancer listener, which holds exactly one key pair, so all
# apps must share one certificate referenced by OCID - and the OCID has to exist before the
# apps step renders its Helm values, which rules out issuing it from inside the cluster.
#
# Terraform cannot own it either. oci_certificates_management_certificate validates IMPORTED
# material in CustomizeDiff with certificateConfigDiffStringProvided(), which reads the value
# via diff.GetChange() and rejects anything empty - and an unknown value reads as empty. So a
# certificate produced in the same apply (by acme_certificate, say) always fails plan with
# "cert_chain_pem is required when config_type is IMPORTED". No version guards this with
# NewValueKnown; checked 8.15 through 8.25. Renewal hits the identical check, so it is not
# only a bootstrap problem.
#
# Until that is automated elsewhere, issue the wildcard (DNS-01 against this zone, since
# Let's Encrypt issues wildcards no other way), import it into OCI Certificates, and pass the
# OCID here.
variable "certificate_ocid" {
  description = "OCID of a certificate in OCI Certificates covering *.<this zone> and the zone apex, served by every app behind the native ingress controller. Passed straight through to the certificate_ocid output."
  type        = string
  default     = ""
}
