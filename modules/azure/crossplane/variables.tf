variable "prefix" {
  type = string
}

variable "resource_group_name" {
  type    = string
  default = ""
}

variable "location" {
  type    = string
  default = ""
}

variable "oidc_issuer_url" {
  type = string
}

variable "node_resource_group_id" {
  type        = string
  description = "AKS node resource group (aks node_resource_group_id), Crossplane may only grant Monitoring Reader there"
}

variable "telemetry_key_resource_id" {
  type        = string
  default     = ""
  description = "kms telemetry key (kms telemetry_key_resource_id), the agent storage account may encrypt with it"
}

variable "acr_id" {
  type        = string
  default     = ""
  description = "Registry with AcrPull for Crossplane core"
}

variable "kubernetes_namespace" {
  type    = string
  default = "crossplane-system"
}

variable "kubernetes_service_account" {
  type        = string
  default     = "crossplane-azure"
  description = "ServiceAccount of the Crossplane Azure providers"
}

variable "kubernetes_core_service_account" {
  type        = string
  default     = "crossplane"
  description = "ServiceAccount of Crossplane core"
}

variable "tags" {
  type    = map(string)
  default = {}
}
