variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the certificate authority."
  type        = string
}

variable "create_ca" {
  description = "Create the certificate authority. Set false to plan a deployment that consumes a CA from elsewhere - the certificate_authority_id output is then empty, and modules/oracle/dns brings its own certificate instead."
  type        = bool
  default     = true
}

# An HSM key is not a preference. OCI Certificates refuses software-protected keys for a
# certificate authority ("Certificates doesn't support the use of software-protected keys" -
# Oracle's own wording), and the Console key picker lists HSM asymmetric keys only.
# modules/oracle/kms creates one as its `ca` key and this is wired to it by agent_input.yaml.
variable "ca_key_id" {
  description = "OCID of an HSM-protected asymmetric key for the CA to sign with. Wired from modules/oracle/kms by agent_input.yaml."
  type        = string
  default     = ""
}

variable "ca_name" {
  description = "Name of the certificate authority. Defaults to <prefix>-root-ca-<random suffix>; see name_salt for why the suffix exists."
  type        = string
  default     = ""
}

variable "description" {
  description = "Description of the certificate authority. Defaults to a generated one naming the prefix."
  type        = string
  default     = ""
}

# CA names are unique across the tenancy *including authorities that are only scheduled for
# deletion*, and a CA cannot be deleted on the spot - OCI enforces a 7-day minimum. So a
# teardown followed by a rebuild inside a week collides with its own leftovers, and so does a
# create that failed *after* OCI accepted it: terraform discards the resource but OCI keeps
# it, FAILED, holding the name for the full week. Scheduling the orphan for deletion does not
# release the name either.
#
# Off by default because a deployment that is built once deserves a name a human can read.
# Turn it on for an environment that is torn down and rebuilt repeatedly - the suffix lives in
# terraform state, so a nuke that takes the state bucket with it produces a fresh one on the
# next run, with nothing to bump by hand.
variable "name_salt" {
  description = "Append a random suffix to the CA name, so a rebuild cannot collide with a torn-down CA still holding its name for 7 days. Leave false for a stable, readable name."
  type        = bool
  default     = false
}

# Subject
#
# Only common_name has a default. The rest are null unless a deployment sets them, which
# leaves them out of the request entirely - an internal CA works without them, but a client
# CA that will be imported into browsers and trust stores is worth naming properly, since
# these fields are what a human sees in a certificate viewer.
variable "common_name" {
  description = "Subject common name of the CA. Defaults to \"<prefix> root CA\"."
  type        = string
  default     = ""
}

variable "organization" {
  description = "Subject organization (O) of the CA."
  type        = string
  default     = null
}

variable "organizational_unit" {
  description = "Subject organizational unit (OU) of the CA."
  type        = string
  default     = null
}

variable "country" {
  description = "Subject country (C) of the CA, as a two-letter code."
  type        = string
  default     = null
}

variable "state_or_province_name" {
  description = "Subject state or province (ST) of the CA."
  type        = string
  default     = null
}

variable "locality_name" {
  description = "Subject locality (L) of the CA."
  type        = string
  default     = null
}

# Validity and issuance rules
variable "ca_validity_years" {
  description = "Lifetime of the CA. Replacing it means redistributing it to every client that trusts it, so this is deliberately long."
  type        = number
  default     = 10
}

variable "leaf_certificate_max_validity" {
  description = "Longest validity this CA will issue a certificate for, as an ISO 8601 duration. Must comfortably exceed the validity of the certificates it signs (OCI's default for those is three months)."
  type        = string
  default     = "P365D"
}

variable "subordinate_ca_max_validity" {
  description = "Longest validity this CA will issue a subordinate CA for, as an ISO 8601 duration. Set so the issuance rule is complete even where nothing creates one."
  type        = string
  default     = "P1095D"
}

variable "signing_algorithm" {
  description = "Signing algorithm for the CA, e.g. SHA256_WITH_RSA. Null leaves OCI to pick one appropriate to ca_key_id."
  type        = string
  default     = null
}

# Setting this makes the CA subordinate instead of self-signed, which is how a deployment
# chains to a root CA held centrally: the root lives in another compartment or tenancy, this
# module creates an intermediate in the deployment's own compartment, and clients that
# already trust the root need no new import.
#
# UNTESTED. Every CA this repo has created so far has been a self-signed root. The issuing
# CA must also grant this compartment the use of it, which is not something this module can
# do from the subordinate's side.
variable "issuer_certificate_authority_id" {
  description = "OCID of a CA to sign this one, making it a subordinate rather than a self-signed root. Leave empty for a root CA. Untested - see the README."
  type        = string
  default     = ""
}

# IAM
#
# Five minutes is not superstition. 60s was tried first and the CA still failed; the same CA,
# created by hand from the same key a few minutes later with nothing else changed, came up
# ACTIVE. So the grant is right and only its propagation is slow. Erring long costs one wait
# on the first deployment; erring short costs a CA stuck in FAILED that OCI will not delete
# for 7 days.
variable "ca_policy_wait" {
  description = "How long to wait after granting certificate authorities the use of this compartment's keys, before creating one. A CA that starts before the grant propagates fails permanently rather than retrying."
  type        = string
  default     = "300s"
}

variable "create_ca_policy" {
  description = "Create the dynamic group and policy that let certificate authorities in this compartment use its keys. Set false only if an equivalent grant already exists - without one, CA creation fails permanently."
  type        = bool
  default     = true
}
