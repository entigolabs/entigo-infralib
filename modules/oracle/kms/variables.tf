variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the vault and its keys."
  type        = string
}

variable "create_vault" {
  description = "Create a vault for these keys instead of using the one vault_id names."
  type        = bool
  default     = false
}

# An OCID rather than a display name: names are not unique in KMS - a compartment can hold
# several vaults called the same thing, including ones only scheduled for deletion - so a
# lookup by name has no single right answer. The agent supplies its own vault's OCID through
# agent_input.yaml, which is how a deployment shares one vault across every module rather
# than spending a slot against the tenancy-wide limit of ten per region.
variable "vault_id" {
  description = "OCID of the vault to place the keys in. Required when create_vault is false, ignored otherwise. Wired from the agent by agent_input.yaml."
  type        = string
  default     = ""
}

variable "vault_name" {
  description = "Display name for the vault this module creates. Defaults to <prefix>-<random suffix>. Ignored when create_vault is false, which takes vault_id instead."
  type        = string
  default     = ""
}

# DEFAULT vaults store keys in OCI's shared HSM partitions and cost nothing for the vault
# itself - you pay per HSM-protected key version. VIRTUAL_PRIVATE gives you dedicated
# partitions and is billed by the hour whether or not it holds any keys, so it is not a
# default anyone should get by accident.
variable "vault_type" {
  description = "DEFAULT or VIRTUAL_PRIVATE. VIRTUAL_PRIVATE is billed per hour - only use it when dedicated HSM partitions are actually required."
  type        = string
  default     = "DEFAULT"
}

# Software-protected key versions are free; HSM-protected ones are billed per version.
# The three storage keys default to SOFTWARE, matching modules/google/kms's
# key_protection_level. The CA key below cannot follow that default - see ca_key_protection_mode.
# A new vault's management endpoint is a hostname of its own
# (<prefix>-management.kms.<region>.oraclecloud.com) and the DNS record for it does not
# exist the moment CreateVault returns. Terraform goes straight on to the keys and every
# one of them fails with "no such host" - seen on the first real run, all four at once,
# right after the vault reported complete after 2m9s. There is nothing to poll and the
# provider does not retry it, so the only fix is to wait.
variable "vault_endpoint_wait" {
  description = "How long to wait after creating a vault before using its management endpoint, so its DNS record can appear. Only applies when this module creates the vault."
  type        = string
  default     = "180s"
}

variable "create_keys" {
  description = "Create the data/config/telemetry/ca keys. Set false to adopt existing keys named by data_key_name/config_key_name/telemetry_key_name/ca_key_name instead of creating a fresh (randomly suffixed) set on every apply."
  type        = bool
  default     = true
}

variable "data_key_name" {
  description = "Display name of the data key. Defaults to <prefix>-data-<random suffix> when creating; when create_keys is false this must name an existing ENABLED key in the vault."
  type        = string
  default     = ""
}

variable "config_key_name" {
  description = "Display name of the config key. Defaults to <prefix>-config-<random suffix> when creating; when create_keys is false this must name an existing ENABLED key in the vault."
  type        = string
  default     = ""
}

variable "telemetry_key_name" {
  description = "Display name of the telemetry key. Defaults to <prefix>-telemetry-<random suffix> when creating; when create_keys is false this must name an existing ENABLED key in the vault."
  type        = string
  default     = ""
}

variable "create_ca_key" {
  description = "Create the asymmetric key that modules/oracle/pca's certificate authority signs with. Set false if nothing in the deployment issues certificates from an OCI CA."
  type        = bool
  default     = true
}

variable "ca_key_name" {
  description = "Display name of the CA signing key. Defaults to <prefix>-ca-<random suffix> when creating; when create_keys is false and create_ca_key is true, this must name an existing ENABLED key in the vault."
  type        = string
  default     = ""
}

# OCI Certificates will not accept a software-protected key for a certificate authority -
# the Console key picker lists HSM-protected asymmetric keys only, and the docs say so
# outright ("Certificates doesn't support the use of software-protected keys"). This is
# the one key in the module that costs money, and it is a per-key-version charge in the
# shared HSM, not a Virtual Private Vault.
variable "ca_key_protection_mode" {
  description = "Protection mode for the CA signing key. OCI Certificates rejects SOFTWARE keys, so changing this away from HSM will break certificate authority creation."
  type        = string
  default     = "HSM"
}

# OCI Certificates accepts RSA 2048/4096 or ECDSA NIST_P384 for a CA. For ECDSA the length
# pairs with a curve - 32 is P-256, 48 is P-384, 66 is P-521 - and CreateKey needs both, so
# main.tf derives curve_id from the length.
variable "ca_key_algorithm" {
  description = "Algorithm for the CA signing key: ECDSA or RSA. OCI Certificates takes ECDSA on NIST P-384 only, so ECDSA means ca_key_length = 48."
  type        = string
  default     = "ECDSA"

  validation {
    condition     = contains(["ECDSA", "RSA"], var.ca_key_algorithm)
    error_message = "ca_key_algorithm must be ECDSA or RSA."
  }
}

variable "ca_key_length" {
  description = "CA signing key length in BYTES: 48 for ECDSA P-384, or 256/512 for RSA-2048/RSA-4096."
  type        = number
  default     = 48
}

variable "key_protection_mode" {
  description = "SOFTWARE or HSM, for the data, config and telemetry keys."
  type        = string
  default     = "SOFTWARE"
}

# These three are master encryption keys - etcd, boot and block volumes, and buckets wrap their
# data keys with them. ECDSA is not an option: it signs, it has no encryption operation at all.
variable "key_algorithm" {
  description = "Algorithm for the data, config and telemetry keys: AES or RSA."
  type        = string
  default     = "AES"

  validation {
    condition     = contains(["AES", "RSA"], var.key_algorithm)
    error_message = "key_algorithm must be AES or RSA. ECDSA cannot encrypt, so it cannot serve as a master encryption key."
  }
}

# NB: OCI expresses key length in BYTES, not bits - 32 is AES-256. RSA takes 256/384/512
# (2048/3072/4096 bits).
variable "key_length" {
  description = "Key length in BYTES for the data, config and telemetry keys. 16, 24 or 32 for AES; 256, 384 or 512 for RSA."
  type        = number
  default     = 32

  validation {
    condition     = contains([16, 24, 32, 256, 384, 512], var.key_length)
    error_message = "key_length must be 16, 24 or 32 (AES) or 256, 384 or 512 (RSA). It is expressed in bytes, not bits."
  }
}

variable "key_rotation_interval_in_days" {
  description = "Rotate the data, config and telemetry keys automatically on this interval. Null leaves rotation off, which is the OCI default and matches aws/kms's enable_key_rotation = false."
  type        = number
  default     = null
}

# An OCI service encrypting something with one of these keys does so as *itself*, and is
# refused unless a policy says otherwise - the same shape of failure the certificate authority
# hits in modules/oracle/pca, and just as unhelpful: nothing mentions a key.
#
# One toggle per consumer rather than a list of service names, because the verb and
# resource-type differ. See the statements in main.tf for which is which.
variable "grant_oke" {
  description = "Let OKE use these keys: \"use keys\" for etcd encryption, plus the key-delegates pair that worker node boot volumes need."
  type        = bool
  default     = true
}

variable "grant_block_storage" {
  description = "Let Block Volume use these keys, which boot and block volumes need."
  type        = bool
  default     = true
}

variable "grant_object_storage" {
  description = "Let Object Storage use these keys, which a bucket with a customer-managed key needs. Requires region, because that principal is region-scoped."
  type        = bool
  default     = true
}

# Escape hatch for a consumer this module does not know about - a database service, say. Given
# verbatim, so the caller owns the verb, resource-type and any conditions.
variable "extra_key_statements" {
  description = "Additional policy statements appended to the key-services policy, verbatim."
  type        = list(string)
  default     = []
}

variable "region" {
  description = "Region whose Object Storage principal is granted key use, e.g. eu-frankfurt-1. Wired from the agent by agent_input.yaml."
  type        = string
  default     = ""
}

# Same reasoning as ca_policy_wait in modules/oracle/pca: IAM is eventually consistent, and a
# resource created before its grant lands fails rather than retrying.
variable "key_policy_wait" {
  description = "How long to wait after granting the service principals key use, before anything encrypts with these keys."
  type        = string
  default     = "60s"
}
