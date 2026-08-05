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

variable "create_certificate" {
  description = "Issue a Let's Encrypt wildcard certificate for the zone and import it into OCI Certificates. Required by modules/k8s/oci-native-ingress-controller, which cannot do SNI and so serves every host from one certificate. Turn off only if nothing on the cluster is exposed through NIC - the ACME DNS-01 challenge needs the zone's NS delegation to be live in the parent domain, which is a manual step, so a brand new zone will fail here until that is done."
  type        = bool
  default     = true
}

variable "acme_email" {
  description = "Contact address registered with Let's Encrypt, used for expiry warnings."
  type        = string
  default     = "tarmo.trumm@entigo.com"
}

variable "certificate_min_days_remaining" {
  description = "Renew the certificate once it has fewer than this many days left. Renewal only happens while terraform is running, so the deployment has to be applied at least this often."
  type        = number
  default     = 30
}

variable "acme_dns_challenge_config" {
  description = "Extra lego configuration for the DNS-01 challenge, merged over OCI_COMPARTMENT_OCID. Normally left empty so credentials come from the environment - see https://go-acme.github.io/lego/dns/oraclecloud/ for the accepted keys."
  type        = map(string)
  default     = {}
  sensitive   = true
}
