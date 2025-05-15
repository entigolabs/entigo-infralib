# Remediate iam-password-policy
resource "aws_iam_account_password_policy" "aws_config" {
  count                          = var.iam_password_policy_enabled ? 1 : 0
  minimum_password_length        = 14
  require_lowercase_characters   = true
  require_numbers                = true
  require_uppercase_characters   = true
  require_symbols                = true
  allow_users_to_change_password = true
  max_password_age               = 90
  password_reuse_prevention      = 24
}

# Remediate multi-region-cloudtrail-enabled
resource "aws_cloudtrail" "aws_config_cloudtrail" {
  count      = var.multi_region_cloudtrail_enabled ? 1 : 0
  depends_on = [aws_s3_bucket_policy.aws_config_cloudtrail[0]]
  tags = {
    created-by = "entigo-infralib"
  }

  name                          = var.prefix
  s3_bucket_name                = aws_s3_bucket.aws_config_cloudtrail[0].id
  s3_key_prefix                 = "prefix"
  include_global_service_events = true
  is_multi_region_trail         = true
}

resource "aws_s3_bucket" "aws_config_cloudtrail" {
  count = var.multi_region_cloudtrail_enabled ? 1 : 0
  bucket = substr(var.cloudtrail_logs_bucket, 0, 63)
  tags = {
    created-by = "entigo-infralib"
  }
}

resource "aws_s3_bucket_policy" "aws_config_cloudtrail" {
  count  = var.multi_region_cloudtrail_enabled ? 1 : 0
  bucket = aws_s3_bucket.aws_config_cloudtrail[0].id
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "AWSCloudTrailAclCheck",
        Effect = "Allow",
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        },
        Action   = "s3:GetBucketAcl",
        Resource = aws_s3_bucket.aws_config_cloudtrail[0].arn
      },
      {
        Sid    = "AWSCloudTrailWrite",
        Effect = "Allow",
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        },
        Action   = "s3:PutObject",
        Resource = "${aws_s3_bucket.aws_config_cloudtrail[0].arn}/prefix/AWSLogs/${data.aws_caller_identity.current.account_id}/*",
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      }
    ]
  })
}
