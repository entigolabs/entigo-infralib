#S3 bucket for VPC flow logs, created when flow_log_destination_type is s3 and no flow_log_s3_destination_arn is given.
resource "aws_s3_bucket" "flow_log" {
  count  = var.enable_flow_log && var.flow_log_destination_type == "s3" && var.flow_log_s3_destination_arn == "" ? 1 : 0
  bucket = trim(substr(lower("${var.prefix}-${data.aws_caller_identity.current.account_id}-${data.aws_region.current.region}-flow-logs"), 0, 63), "-")
  #Deleting the bucket fails while it still has logs, empty it first
  force_destroy = false
  tags = {
    created-by = "entigo-infralib"
  }
}

resource "aws_s3_bucket_public_access_block" "flow_log" {
  count  = length(aws_s3_bucket.flow_log)
  bucket = aws_s3_bucket.flow_log[0].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "flow_log" {
  count  = length(aws_s3_bucket.flow_log)
  bucket = aws_s3_bucket.flow_log[0].id

  rule {
    id     = "expire-flow-logs"
    status = "Enabled"
    filter {}
    expiration {
      days = var.flow_log_s3_retention_in_days
    }
    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

#KMS key when flow_log_s3_kms_key_arn is set, otherwise SSE-S3. The key policy must allow delivery.logs.amazonaws.com
#https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs-s3-cmk-policy.html
resource "aws_s3_bucket_server_side_encryption_configuration" "flow_log" {
  count  = length(aws_s3_bucket.flow_log)
  bucket = aws_s3_bucket.flow_log[0].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = var.flow_log_s3_kms_key_arn != "" ? "aws:kms" : "AES256"
      kms_master_key_id = var.flow_log_s3_kms_key_arn != "" ? var.flow_log_s3_kms_key_arn : null
    }
    bucket_key_enabled       = var.flow_log_s3_kms_key_arn != ""
    blocked_encryption_types = ["SSE-C"]
  }
}

#https://docs.aws.amazon.com/vpc/latest/userguide/flow-logs-s3-permissions.html
resource "aws_s3_bucket_policy" "flow_log" {
  count  = length(aws_s3_bucket.flow_log)
  bucket = aws_s3_bucket.flow_log[0].id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AWSLogDeliveryWrite"
        Effect    = "Allow"
        Principal = { Service = "delivery.logs.amazonaws.com" }
        Action    = "s3:PutObject"
        Resource  = "${aws_s3_bucket.flow_log[0].arn}/AWSLogs/${data.aws_caller_identity.current.account_id}/*"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
            "s3:x-amz-acl"      = "bucket-owner-full-control"
          }
          ArnLike = {
            "aws:SourceArn" = "arn:aws:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:*"
          }
        }
      },
      {
        Sid       = "AWSLogDeliveryAclCheck"
        Effect    = "Allow"
        Principal = { Service = "delivery.logs.amazonaws.com" }
        Action    = ["s3:GetBucketAcl", "s3:ListBucket"]
        Resource  = aws_s3_bucket.flow_log[0].arn
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:aws:logs:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:*"
          }
        }
      },
      {
        Sid       = "DenyInsecureTransport"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource  = [aws_s3_bucket.flow_log[0].arn, "${aws_s3_bucket.flow_log[0].arn}/*"]
        Condition = {
          Bool = { "aws:SecureTransport" = "false" }
        }
      }
    ]
  })
  depends_on = [aws_s3_bucket_public_access_block.flow_log, aws_s3_bucket_server_side_encryption_configuration.flow_log]
}
