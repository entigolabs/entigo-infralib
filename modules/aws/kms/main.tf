
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
  key_owners                             = [data.aws_iam_session_context.current.issuer_arn]
  
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
          values = concat([
            "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:${var.policy_prefix}",
            "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/eks/${var.policy_prefix}",
          ], var.telemetry_extra_encryption_context)
        }
      ]
    },
    {
      principals = [
        {
          type        = "AWS"
          identifiers = ["*"]
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
          test     = "StringLike"
          variable = "aws:PrincipalArn"
          values = [for role in var.telemetry_extra_bucket_roles : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${role}"]
        }
      ]
    },
    {
      principals = [
        {
          type        = "AWS"
          identifiers = ["*"]
        }
      ]
    
      actions = [      
        "kms:CreateGrant",
        "kms:ListGrants",
        "kms:RevokeGrant"
      ]

      resources = [
        "*",
      ]
      
      conditions = [
        {
          test     = "StringLike"
          variable = "aws:PrincipalArn"
          values = [for role in var.telemetry_extra_bucket_roles : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${role}"]
        },
        {
          test     = "Bool"
          variable = "kms:GrantIsForAWSResource"
          values = [
            "true"
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
  key_owners                             = [data.aws_iam_session_context.current.issuer_arn]
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
  key_owners                             = [data.aws_iam_session_context.current.issuer_arn]
  key_service_roles_for_autoscaling      = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"]
  
  key_statements = [
    {
      principals = [
        {
          type        = "AWS"
          identifiers = ["*"]
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
          test     = "StringLike"
          variable = "aws:PrincipalArn"
          values = [
            "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.policy_prefix}-ebs-csi"
          ]
        }
      ]
    },
    {
      principals = [
        {
          type        = "AWS"
          identifiers = ["*"]
        }
      ]
    
      actions = [      
        "kms:CreateGrant",
        "kms:ListGrants",
        "kms:RevokeGrant"
      ]

      resources = [
        "*",
      ]
      
      conditions = [
        {
          test     = "StringLike"
          variable = "aws:PrincipalArn"
          values = [
            "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.policy_prefix}-ebs-csi"
          ]
        },
        {
          test     = "Bool"
          variable = "kms:GrantIsForAWSResource"
          values = [
            "true"
          ]
        }
      ]
    }
  ]
  
  
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
