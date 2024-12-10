data "aws_iam_roles" "aws-admin-roles" {
  count = var.iam_admin_role != "" ? 1 : 0
  name_regex  = var.iam_admin_role
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
