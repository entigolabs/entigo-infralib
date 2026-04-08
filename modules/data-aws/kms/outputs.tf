output "telemetry_alias_arn" {
  value = length(data.aws_kms_alias.telemetry) > 0 ? data.aws_kms_alias.telemetry[0].arn : null
}

output "telemetry_key_arn" {
  value = data.aws_kms_key.telemetry.arn
}

output "telemetry_key_id" {
  value = data.aws_kms_key.telemetry.key_id
}

output "telemetry_key_policy" {
  value = null
}

output "config_alias_arn" {
  value = length(data.aws_kms_alias.config) > 0 ? data.aws_kms_alias.config[0].arn : null
}

output "config_key_arn" {
  value = data.aws_kms_key.config.arn
}

output "config_key_id" {
  value = data.aws_kms_key.config.key_id
}

output "config_key_policy" {
  value = null
}

output "data_alias_arn" {
  value = length(data.aws_kms_alias.data) > 0 ? data.aws_kms_alias.data[0].arn : null
}

output "data_key_arn" {
  value = data.aws_kms_key.data.arn
}

output "data_key_id" {
  value = data.aws_kms_key.data.key_id
}

output "data_key_policy" {
  value = null
}
