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
