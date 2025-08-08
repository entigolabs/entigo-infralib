variable "prefix" {
  type = string
}

variable "eks_oidc_provider" {
  type = string
}

variable "eks_oidc_provider_arn" {
  type = string
}


variable "kubernetes_service_account" {
  type = string
  description = "Kubernetes service account name for aws crossplane provider"
  default = "crossplane-aws"
}

variable "kubernetes_namespace" {
  type = string
  description = "Kubernetes namespace name for crossplane"
  default = "crossplane-system"
}

#variable "ecr_proxy_policy_arn" {
#  type = string
#  description = "Policy to attatch to crossplane core for ecr access"
#  default = ""
#}

#variable "kubernetes_core_service_account" {
#  type = string
#  description = "Kubernetes service account name for crossplane core"
#  default = "crossplane"
#}
