data "aws_ssm_parameter" "eks_oidc_provider" {
  name = "/entigo-infralib/${var.eks_prefix}-${terraform.workspace}/eks/oidc_provider"
}

data "aws_ssm_parameter" "eks_oidc_provider_arn" {
  name = "/entigo-infralib/${var.eks_prefix}-${terraform.workspace}/eks/oidc_provider_arn"
}

data "aws_ssm_parameter" "region" {
  name = "/entigo-infralib/${var.eks_prefix}-${terraform.workspace}/eks/region"
}

data "aws_ssm_parameter" "account" {
  name = "/entigo-infralib/${var.eks_prefix}-${terraform.workspace}/eks/account"
}
