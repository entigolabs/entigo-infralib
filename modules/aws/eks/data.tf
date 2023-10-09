data "aws_iam_roles" "aws-admin-roles" {
  name_regex  = var.iam_admin_role
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
