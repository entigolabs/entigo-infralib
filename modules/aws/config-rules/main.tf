resource "aws_config_aggregate_authorization" "config_rules" {
  count                 = (var.aggregate_authorization_account_id != "" && var.aggregate_authorization_authorized_aws_region != "") ? 1 : 0
  account_id            = var.aggregate_authorization_account_id
  region                = var.aggregate_authorization_authorized_aws_region
  # Starting from provider version 6.0.0 "region" is deprecated and "authorized_aws_region" must be used
  # https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/config_aggregate_authorization
  # authorized_aws_region = var.aggregate_authorization_authorized_aws_region
  tags = {
    created-by = "entigo-infralib"
  }
}

resource "aws_config_configuration_recorder" "config_rules" {
  name     = var.prefix
  role_arn = aws_iam_role.config_rules.arn

  recording_group {
    all_supported                 = length(var.resource_types_to_exclude) == 0 ? true : false
    include_global_resource_types = length(var.resource_types_to_exclude) == 0 ? true : false

    dynamic "exclusion_by_resource_types" {
      for_each = length(var.resource_types_to_exclude) > 0 ? [1] : []
      content {
        resource_types = var.resource_types_to_exclude
      }
    }

    recording_strategy {
      use_only = length(var.resource_types_to_exclude) == 0 ? "ALL_SUPPORTED_RESOURCE_TYPES" : "EXCLUSION_BY_RESOURCE_TYPES"
    }
  }
}

resource "aws_config_delivery_channel" "config_rules" {
  name           = var.prefix
  s3_bucket_name = aws_s3_bucket.config_rules_logs.id
  s3_key_prefix  = "config-logs"

  snapshot_delivery_properties {
    delivery_frequency = "TwentyFour_Hours"
  }

  depends_on = [aws_config_configuration_recorder.config_rules]
}

resource "aws_config_configuration_recorder_status" "config_rules" {
  name       = aws_config_configuration_recorder.config_rules.name
  is_enabled = true
  depends_on = [aws_config_delivery_channel.config_rules]
}

# IAM Role for AWS Config
resource "aws_iam_role" "config_rules" {
  name = var.prefix
  tags = {
    created-by = "entigo-infralib"
  }

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "config_rules" {
  role       = aws_iam_role.config_rules.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWS_ConfigRole"
}

# S3 Bucket for Config logs
resource "aws_s3_bucket" "config_rules_logs" {
  bucket = substr(var.config_logs_bucket, 0, 63)
  tags = {
    created-by = "entigo-infralib"
  }
}

resource "aws_s3_bucket_versioning" "config_rules_logs" {
  bucket = aws_s3_bucket.config_rules_logs.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_public_access_block" "config_rules_logs" {
  bucket = aws_s3_bucket.config_rules_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_policy" "config_rules_logs" {
  bucket = aws_s3_bucket.config_rules_logs.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSConfigBucketPermissionsCheck"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = "arn:aws:s3:::${aws_s3_bucket.config_rules_logs.id}"
      },
      {
        Sid    = "AWSConfigBucketDelivery"
        Effect = "Allow"
        Principal = {
          Service = "config.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "arn:aws:s3:::${aws_s3_bucket.config_rules_logs.id}/config-logs/AWSLogs/${data.aws_caller_identity.current.account_id}/Config/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      }
    ]
  })
}
