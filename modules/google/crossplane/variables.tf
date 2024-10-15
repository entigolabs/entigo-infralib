variable "prefix" {
  type = string
}

variable "kubernetes_service_account" {
  type = string
  description = "Kubernetes service account name for google crossplane provider"
  default = "crossplane-google"
}

variable "kubernetes_namespace" {
  type = string
  description = "Kubernetes namespace name for crossplane"
  default = "crossplane-system"
}
