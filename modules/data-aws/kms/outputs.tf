output "telemetry_alias_arn" {
  value = data.aws_kms_alias.telemetry.arn
}

output "telemetry_key_arn" {
  value = data.aws_kms_key.telemetry.arn
}

output "telemetry_key_id" {
  value = data.aws_kms_key.telemetry.key_id
}

output "telemetry_key_policy" {
  value = data.aws_kms_key.telemetry.policy
}

output "config_alias_arn" {
  value = data.aws_kms_alias.config.arn
}

output "config_key_arn" {
  value = data.aws_kms_key.config.arn
}

output "config_key_id" {
  value = data.aws_kms_key.config.key_id
}

output "config_key_policy" {
  value = data.aws_kms_key.config.policy
}

output "data_alias_arn" {
  value = data.aws_kms_alias.data.arn
}

output "data_key_arn" {
  value = data.aws_kms_key.data.arn
}

output "data_key_id" {
  value = data.aws_kms_key.data.key_id
}

output "data_key_policy" {
  value = data.aws_kms_key.data.policy
}
