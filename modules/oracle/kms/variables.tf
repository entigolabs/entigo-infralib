variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the vault and its keys."
  type        = string
}

variable "create_vault" {
  description = "Create the vault. Set false to place the keys in an existing vault named by vault_name."
  type        = bool
  default     = true
}

variable "vault_name" {
  description = "Display name of the vault. Defaults to <prefix>-<random suffix> when creating; when create_vault is false this must name an existing vault in compartment_id."
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

variable "key_protection_mode" {
  description = "SOFTWARE or HSM, for the data, config and telemetry keys."
  type        = string
  default     = "SOFTWARE"
}

variable "key_algorithm" {
  description = "Key algorithm for the data, config and telemetry keys: AES, RSA or ECDSA."
  type        = string
  default     = "AES"
}

# NB: OCI expresses key length in BYTES, not bits - 32 is AES-256. RSA takes 256/384/512
# (2048/3072/4096 bits).
variable "key_length" {
  description = "Key length in BYTES for the data, config and telemetry keys. 16, 24 or 32 for AES; 256, 384 or 512 for RSA."
  type        = number
  default     = 32
}

variable "key_rotation_interval_in_days" {
  description = "Rotate the data, config and telemetry keys automatically on this interval. Null leaves rotation off, which is the OCI default and matches aws/kms's enable_key_rotation = false."
  type        = number
  default     = null
}

variable "create_ca_key" {
  description = "Create the asymmetric key that modules/oracle/pca's certificate authority signs with. Set false if nothing in the deployment issues certificates from an OCI CA."
  type        = bool
  default     = true
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

# OCI Certificates accepts RSA 2048/4096 or ECDSA NIST_P384 for a CA. Only RSA is wired
# here: key_shape.length is required by the provider and is meaningless for ECDSA, so
# supporting both would mean a branch that cannot be exercised without a live ECDSA CA.
variable "ca_key_algorithm" {
  description = "Algorithm for the CA signing key. RSA only - OCI Certificates also allows ECDSA NIST_P384, but this module does not wire the curve."
  type        = string
  default     = "RSA"
}

variable "ca_key_length" {
  description = "CA signing key length in BYTES: 256 for RSA-2048, 512 for RSA-4096."
  type        = number
  default     = 512
}
