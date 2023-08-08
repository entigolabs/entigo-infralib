data "aws_ssm_parameter" "private_subnet_cidrs" {
  name = "/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/vpc/private_subnet_cidrs"
}

#data "aws_ssm_parameter" "intra_subnet_cidrs" {
#  name = "/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/vpc/intra_subnet_cidrs"
#}

data "aws_ssm_parameter" "vpc_id" {
  name = "/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/vpc/vpc_id"
}

data "aws_ssm_parameter" "private_subnets" {
  name = "/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/vpc/private_subnets"
}

data "aws_ssm_parameter" "public_subnets" {
  name = "/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/vpc/public_subnets"
}

data "aws_ssm_parameters_by_path" "vpc" {
  path      = "/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/vpc"
  recursive = true
}

data "aws_iam_roles" "aws-admin-roles" {
  name_regex  = var.iam_admin_role
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
