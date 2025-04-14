module "aws_config" {
  source  = "cloudposse/config/aws"
  version = "0.13.0"

  name = "aws-config"

  # Configure AWS Config to deliver snapshots once per day
  delivery_frequency = "TwentyFour_Hours"
  
  # S3 bucket for storing logs
  s3_bucket_enabled = true
  s3_bucket_name    = "aws-config-logs-${data.aws_caller_identity.current.account_id}"
  s3_key_prefix     = "config-logs"

  # Recording group configuration
  recording_group_all_supported                 = true
  recording_group_include_global_resource_types = true

  # Configure SNS topic for notifications (optional)
  sns_topic_enabled = false
  # sns_topic_name  = "aws-config-notifications"

  # Optional KMS key for encryption
  # kms_key_enabled = true
  # kms_key_alias   = "alias/aws-config-encryption"
}