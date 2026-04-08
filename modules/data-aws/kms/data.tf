data "aws_kms_alias" "telemetry" {
  count = var.telemetry_key_id == null ? 1 : 0
  name  = "alias/${var.prefix}/telemetry"
}

data "aws_kms_alias" "config" {
  count = var.config_key_id == null ? 1 : 0
  name  = "alias/${var.prefix}/config"
}

data "aws_kms_alias" "data" {
  count = var.data_key_id == null ? 1 : 0
  name  = "alias/${var.prefix}/data"
}

locals {
  telemetry_key_id = var.telemetry_key_id != null ? var.telemetry_key_id : data.aws_kms_alias.telemetry[0].target_key_id
  config_key_id    = var.config_key_id != null ? var.config_key_id : data.aws_kms_alias.config[0].target_key_id
  data_key_id      = var.data_key_id != null ? var.data_key_id : data.aws_kms_alias.data[0].target_key_id
}

data "aws_kms_key" "telemetry" {
  key_id = local.telemetry_key_id
}

data "aws_kms_key" "config" {
  key_id = local.config_key_id
}

data "aws_kms_key" "data" {
  key_id = local.data_key_id
}
