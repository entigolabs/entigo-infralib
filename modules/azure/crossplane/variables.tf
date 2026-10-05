variable "prefix" {
  type = string
}

variable "resource_group_name" {
  type = string
}

variable "location" {
  type = string
}

variable "oidc_issuer_url" {
  type = string
}

variable "node_resource_group_id" {
  type = string
}

variable "acr_id" {
  type    = string
  default = ""
}

variable "kubernetes_namespace" {
  type    = string
  default = "crossplane-system"
}

variable "kubernetes_service_account" {
  type    = string
  default = "crossplane-azure"
}

variable "kubernetes_core_service_account" {
  type    = string
  default = "crossplane"
}

variable "tags" {
  type    = map(string)
  default = {}
}
