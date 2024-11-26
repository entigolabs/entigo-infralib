
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
  
  key_statements = [
    {
      principals = [
        {
          type        = "Service"
          identifiers = ["logs.${data.aws_region.current.name}.amazonaws.com"]
        }
      ]
    
      actions = [      
        "kms:Encrypt*",
        "kms:Decrypt*",
        "kms:ReEncrypt*",
        "kms:GenerateDataKey*",
        "kms:Describe*"
      ]

      resources = [
        "*",
      ]
      
      conditions = [
        {
          test     = "ArnLike"
          variable = "kms:EncryptionContext:aws:logs:arn"
          values = [
            "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:${var.prefix}/*",
            "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/eks/${var.prefix}/*",
          ]
        }
      ]
      
    }
  ]
  
  
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
  key_service_roles_for_autoscaling      = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"]
  aliases = ["${var.prefix}/data"]

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}


resource "aws_ssm_parameter" "telemetry_alias_arn" {
  count = var.mode == "kms" ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/telemetry_alias_arn"
  type  = "String"
  value = module.kms_telemetry[0].aliases["${var.prefix}/telemetry"].arn
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "config_alias_arn" {
  count = var.mode == "kms" ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/config_alias_arn"
  type  = "String"
  value = module.kms_config[0].aliases["${var.prefix}/config"].arn
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_ssm_parameter" "data_alias_arn" {
  count = var.mode == "kms" ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/data_alias_arn"
  type  = "String"
  value = module.kms_data[0].aliases["${var.prefix}/data"].arn
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}
