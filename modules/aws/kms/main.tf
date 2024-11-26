
#https://registry.terraform.io/modules/terraform-aws-modules/kms/aws/latest

module "kms_telemetry" {
  source = "terraform-aws-modules/kms/aws"
  version = "3.1.1"
  count = var.mode == "kms" ? 1 : 0
  deletion_window_in_days = var.deletion_window_in_days
  description             = "${var.prefix} telemetry"
  enable_key_rotation     = var.enable_key_rotation
  is_enabled              = true
  key_usage               = "ENCRYPT_DECRYPT"
  multi_region            = var.multi_region
  enable_default_policy                  = true
  key_owners                             = [data.aws_caller_identity.current.arn]
  key_administrators                     = [data.aws_caller_identity.current.arn]
  key_users                              = [data.aws_caller_identity.current.arn]
  key_service_users                      = [data.aws_caller_identity.current.arn]
  key_service_roles_for_autoscaling      = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"]
  key_symmetric_encryption_users         = [data.aws_caller_identity.current.arn]
  key_hmac_users                         = [data.aws_caller_identity.current.arn]
  key_asymmetric_public_encryption_users = [data.aws_caller_identity.current.arn]
  key_asymmetric_sign_verify_users       = [data.aws_caller_identity.current.arn]
  aliases = ["${var.prefix}/telemetry"]

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

module "kms_config" {
  source = "terraform-aws-modules/kms/aws"
  version = "3.1.1"
  count = var.mode == "kms" ? 1 : 0
  deletion_window_in_days = var.deletion_window_in_days
  description             = "${var.prefix} config"
  enable_key_rotation     = var.enable_key_rotation
  is_enabled              = true
  key_usage               = "ENCRYPT_DECRYPT"
  multi_region            = var.multi_region
  enable_default_policy                  = true
  key_owners                             = [data.aws_caller_identity.current.arn]
  key_administrators                     = [data.aws_caller_identity.current.arn]
  key_users                              = [data.aws_caller_identity.current.arn]
  key_service_users                      = [data.aws_caller_identity.current.arn]
  key_service_roles_for_autoscaling      = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"]
  key_symmetric_encryption_users         = [data.aws_caller_identity.current.arn]
  key_hmac_users                         = [data.aws_caller_identity.current.arn]
  key_asymmetric_public_encryption_users = [data.aws_caller_identity.current.arn]
  key_asymmetric_sign_verify_users       = [data.aws_caller_identity.current.arn]
  aliases = ["${var.prefix}/config"]

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

module "kms_data" {
  source = "terraform-aws-modules/kms/aws"
  version = "3.1.1"
  count = var.mode == "kms" ? 1 : 0
  deletion_window_in_days = var.deletion_window_in_days
  description             = "${var.prefix} data"
  enable_key_rotation     = var.enable_key_rotation
  is_enabled              = true
  key_usage               = "ENCRYPT_DECRYPT"
  multi_region            = var.multi_region
  enable_default_policy                  = true
  key_owners                             = [data.aws_caller_identity.current.arn]
  key_administrators                     = [data.aws_caller_identity.current.arn]
  key_users                              = [data.aws_caller_identity.current.arn]
  key_service_users                      = [data.aws_caller_identity.current.arn]
  key_service_roles_for_autoscaling      = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"]
  key_symmetric_encryption_users         = [data.aws_caller_identity.current.arn]
  key_hmac_users                         = [data.aws_caller_identity.current.arn]
  key_asymmetric_public_encryption_users = [data.aws_caller_identity.current.arn]
  key_asymmetric_sign_verify_users       = [data.aws_caller_identity.current.arn]
  aliases = ["${var.prefix}/data"]

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}


resource "aws_ssm_parameter" "telemetry_key_arn" {
  count = var.mode == "kms" ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/telemetry_key_arn"
  type  = "String"
  value = module.kms_telemetry[0].key_arn
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "config_key_arn" {
  count = var.mode == "kms" ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/config_key_arn"
  type  = "String"
  value = module.kms_config[0].key_arn
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "data_key_arn" {
  count = var.mode == "kms" ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/data_key_arn"
  type  = "String"
  value = module.kms_data[0].key_arn
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}
