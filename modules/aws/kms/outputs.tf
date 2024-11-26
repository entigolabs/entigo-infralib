output "telemetry_aliases" {
  value = var.mode == "kms" ? module.kms_telemetry[0].aliases : null   
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

output "config_aliases" {
  value = var.mode == "kms" ? module.kms_config[0].aliases : null   
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

output "data_aliases" {
  value = var.mode == "kms" ? module.kms_data[0].aliases : null   
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

output "deletion_window_in_days" {
  value = var.deletion_window_in_days
}

output "enable_key_rotation" {
  value = var.enable_key_rotation
}

output "multi_region" {
  value = var.multi_region
}

output "encryption_enabled" {
  value = var.mode == "disabled" ? false : true
}

output "mode" {
  value = var.mode
}
