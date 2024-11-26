output "telemetry_alias_arn" {
  value = var.mode == "kms" ? module.kms_telemetry[0].aliases["${var.prefix}/telemetry"].arn : null
}

output "telemetry_key_arn" {
  value = var.mode == "kms" ? module.kms_telemetry[0].key_arn : null  
}

output "telemetry_key_id" {
  value = var.mode == "kms" ? module.kms_telemetry[0].key_id : null  
}

output "telemetry_key_policy" {
  value = var.mode == "kms" ? module.kms_telemetry[0].key_policy : null  
}

output "config_alias_arn" {
  value = var.mode == "kms" ? module.kms_config[0].aliases["${var.prefix}/config"].arn : null
}

output "config_key_arn" {
  value = var.mode == "kms" ? module.kms_config[0].key_arn : null  
}

output "config_key_id" {
  value = var.mode == "kms" ? module.kms_config[0].key_id : null  
}

output "config_key_policy" {
  value = var.mode == "kms" ? module.kms_config[0].key_policy : null  
}

output "data_alias_arn" {
  value = var.mode == "kms" ? module.kms_data[0].aliases["${var.prefix}/data"].arn : null
}

output "data_key_arn" {
  value = var.mode == "kms" ? module.kms_data[0].key_arn : null  
}

output "data_key_id" {
  value = var.mode == "kms" ? module.kms_data[0].key_id : null  
}

output "data_key_policy" {
  value = var.mode == "kms" ? module.kms_data[0].key_policy : null  
}
