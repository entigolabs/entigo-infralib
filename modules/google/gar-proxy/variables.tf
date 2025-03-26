variable "prefix" {
  type = string
}

variable "enablement_config" {
  description = "This configures whether vulnerability scanning is automatically performed for artifacts pushed to this repository. Possible values are: INHERITED, DISABLED."
  type = string
  default = "DISABLED"
}

variable "hub_username_secret" {
  description = "Secret Manager secret name for Docker Hub username"
  type        = string
  default     = ""
}

variable "hub_access_token_secret" {
  description = "Secret Manager secret name for Docker Hub access token"
  type        = string
  default     = ""
}

variable "ghcr_username_secret" {
  description = "Secret Manager secret name for GitHub Container Registry username"
  type        = string
  default     = ""
}

variable "ghcr_access_token_secret" {
  description = "Secret Manager secret name for GitHub Container Registry access token"
  type        = string
  default     = ""
}

variable "gcr_username_secret" {
  description = "Secret Manager secret name for Google Container Registry username"
  type        = string
  default     = ""
}

variable "gcr_access_token_secret" {
  description = "Secret Manager secret name for Google Container Registry access token"
  type        = string
  default     = ""
}

variable "ecr_username_secret" {
  description = "Secret Manager secret name for Amazon ECR username"
  type        = string
  default     = ""
}

variable "ecr_access_token_secret" {
  description = "Secret Manager secret name for Amazon ECR access token"
  type        = string
  default     = ""
}

variable "quay_username_secret" {
  description = "Secret Manager secret name for Quay username"
  type        = string
  default     = ""
}

variable "quay_access_token_secret" {
  description = "Secret Manager secret name for Quay access token"
  type        = string
  default     = ""
}

