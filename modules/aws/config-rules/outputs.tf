output "config_rules_logs_bucket_name" {
  value = aws_s3_bucket.config_rules_logs.id
}

output "cloudtrail_logs_bucket_name" {
  value = aws_s3_bucket.aws_config_cloudtrail[0].id
}

output "prefix" {
  value = var.prefix
}