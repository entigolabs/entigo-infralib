variable "prefix" {
  type = string
}

variable "crossplane_service_account_id" {
  type        = string
  description = "Google service account ID for Crossplane"
  default     = ""
}

variable "kubernetes_service_account" {
  type        = string
  description = "Kubernetes service account name for Google Crossplane provider"
  default     = "crossplane-google"
}

variable "kubernetes_namespace" {
  type        = string
  description = "Kubernetes namespace name for Crossplane"
  default     = "crossplane-system"
}
