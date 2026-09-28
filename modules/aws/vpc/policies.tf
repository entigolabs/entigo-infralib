locals {
  endpoint_policy_service = {
    s3      = "s3"
    s3e     = "s3"
    ecr_api = "ecr"
    ecr_dkr = "ecr"
    ec2     = "ec2"
    sts     = "sts"
    efs     = "elasticfilesystem"
  }
  # Custom policy, otherwise the default policy scoped to the endpoint service (s3:* instead of *)
  endpoint_policy = {
    for k, service in local.endpoint_policy_service : k => lookup(var.endpoint_policies, k, data.aws_iam_policy_document.endpoint[service].json)
  }
}

data "aws_iam_policy_document" "endpoint" {
  for_each = toset(values(local.endpoint_policy_service))

  statement {
    sid       = "AllowService"
    actions   = ["${each.key}:*"]
    resources = ["*"]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
  }
}
