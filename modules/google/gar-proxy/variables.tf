variable "prefix" {
  type = string
}

variable "hub_username" {
  description = "Docker Hub username"
  type        = string
  default     = ""
}

variable "hub_access_token_secret_name" {
  description = "Docker Hub access token secret name"
  type        = string
  default     = ""
}

variable "ghcr_username" {
  description = "GitHub Container Registry username"
  type        = string
  default     = ""
}

variable "ghcr_access_token_secret_name" {
  description = "GitHub Container Registry access token secret name"
  type        = string
  default     = ""
}

variable "gcr_username" {
  description = "Google Container Registry username"
  type        = string
  default     = ""
}

variable "gcr_access_token_secret_name" {
  description = "Google Container Registry access token secret name"
  type        = string
  default     = ""
}
