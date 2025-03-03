variable "prefix" {
  type = string
}

variable "hub_username" {
  type = string
  sensitive   = true
  default = ""
}

variable "hub_token" {
  type = string
  sensitive   = true
  default = ""
}

variable "ghcr_username" {
  type = string
  sensitive   = true
  default = ""
}

variable "ghcr_token" {
  type = string
  sensitive   = true
  default = ""
}

variable "gcr_username" {
  type = string
  sensitive   = true
  default = ""
}

variable "gcr_token" {
  type = string
  sensitive   = true
  default = ""
}
