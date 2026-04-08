data "aws_kms_alias" "telemetry" {
  name = "alias/${var.prefix}/telemetry"
}

data "aws_kms_alias" "config" {
  name = "alias/${var.prefix}/config"
}

data "aws_kms_alias" "data" {
  name = "alias/${var.prefix}/data"
}

data "aws_kms_key" "telemetry" {
  key_id = data.aws_kms_alias.telemetry.target_key_id
}

data "aws_kms_key" "config" {
  key_id = data.aws_kms_alias.config.target_key_id
}

data "aws_kms_key" "data" {
  key_id = data.aws_kms_alias.data.target_key_id
}
