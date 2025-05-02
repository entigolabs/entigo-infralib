output "bucket_name" {
  value = aws_s3_bucket.config_logs.id
}

output "prefix" {
  value = var.prefix
}