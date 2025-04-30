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

variable "crossplane_core_iam_policy" {
  type = string
  description = "Policy for crossplane-core kubernetes service account"
  default = ""
}
