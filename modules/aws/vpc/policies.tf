data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

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
  # Custom policy, otherwise the default policy
  endpoint_policy = {
    for k, service in local.endpoint_policy_service : k => lookup(var.endpoint_policies, k, data.aws_iam_policy_document.endpoint[service].json)
  }
  account_id = data.aws_caller_identity.current.account_id
  partition  = data.aws_partition.current.partition
  region     = data.aws_region.current.region
}

# Default policy: only the current account, the organization (if set) and AWS services can use the endpoint
data "aws_iam_policy_document" "endpoint" {
  for_each = toset(values(local.endpoint_policy_service))

  statement {
    sid       = "TrustedAccount"
    actions   = ["${each.key}:*"]
    resources = ["*"]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalAccount"
      values   = [local.account_id]
    }
  }

  dynamic "statement" {
    for_each = var.endpoint_policy_org_id != "" ? [1] : []
    content {
      sid       = "TrustedOrganization"
      actions   = ["${each.key}:*"]
      resources = ["*"]
      principals {
        type        = "*"
        identifiers = ["*"]
      }
      condition {
        test     = "StringEquals"
        variable = "aws:PrincipalOrgID"
        values   = [var.endpoint_policy_org_id]
      }
    }
  }

  statement {
    sid       = "AWSServices"
    actions   = ["${each.key}:*"]
    resources = ["*"]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    condition {
      test     = "Bool"
      variable = "aws:PrincipalIsAWSService"
      values   = ["true"]
    }
  }

  # AWS owned buckets are not read with our credentials: ECR image layers (presigned URLs) and Amazon Linux repositories
  dynamic "statement" {
    for_each = each.key == "s3" ? [1] : []
    content {
      sid     = "AWSOwnedBuckets"
      actions = ["s3:GetObject"]
      resources = [
        "arn:${local.partition}:s3:::prod-${local.region}-starport-layer-bucket/*",
        "arn:${local.partition}:s3:::al2023-repos-${local.region}-*/*",
        "arn:${local.partition}:s3:::amazonlinux-2-repos-${local.region}/*",
      ]
      principals {
        type        = "*"
        identifiers = ["*"]
      }
    }
  }

  # AssumeRoleWithWebIdentity (IRSA) has no caller account, so check the account that owns the role
  dynamic "statement" {
    for_each = each.key == "sts" ? [1] : []
    content {
      sid       = "WebIdentity"
      actions   = ["sts:AssumeRoleWithWebIdentity", "sts:AssumeRoleWithSAML"]
      resources = ["*"]
      principals {
        type        = "*"
        identifiers = ["*"]
      }
      condition {
        test     = "StringEquals"
        variable = "aws:ResourceAccount"
        values   = [local.account_id]
      }
    }
  }

  dynamic "statement" {
    for_each = each.key == "sts" && var.endpoint_policy_org_id != "" ? [1] : []
    content {
      sid       = "WebIdentityOrganization"
      actions   = ["sts:AssumeRoleWithWebIdentity", "sts:AssumeRoleWithSAML"]
      resources = ["*"]
      principals {
        type        = "*"
        identifiers = ["*"]
      }
      condition {
        test     = "StringEquals"
        variable = "aws:ResourceOrgID"
        values   = [var.endpoint_policy_org_id]
      }
    }
  }
}
