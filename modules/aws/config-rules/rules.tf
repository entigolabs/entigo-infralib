# AWS Config Rules converted from CloudFormation
# Source: https://raw.githubusercontent.com/awslabs/aws-config-rules/master/aws-config-conformance-packs/Operational-Best-Practices-for-CIS-AWS-v1.4-Level1.yaml

resource "aws_config_config_rule" "accesskeysrotated" {
  name  = "access-keys-rotated"

  source {
    owner             = "AWS"
    source_identifier = "ACCESS_KEYS_ROTATED"
  }

  input_parameters = jsonencode({
    maxAccessKeyAge = 90
  })
}

resource "aws_config_config_rule" "cloudtrailcloudwatchlogsenabled" {
  name  = "cloud-trail-cloud-watch-logs-enabled"

  source {
    owner             = "AWS"
    source_identifier = "CLOUD_TRAIL_CLOUD_WATCH_LOGS_ENABLED"
  }
}

resource "aws_config_config_rule" "ec2ebsencryptionbydefault" {
  name  = "ec2-ebs-encryption-by-default"

  source {
    owner             = "AWS"
    source_identifier = "EC2_EBS_ENCRYPTION_BY_DEFAULT"
  }
}

resource "aws_config_config_rule" "encryptedvolumes" {
  name  = "encrypted-volumes"

  source {
    owner             = "AWS"
    source_identifier = "ENCRYPTED_VOLUMES"
  }
}

resource "aws_config_config_rule" "iamnoinlinepolicycheck" {
  name  = "iam-no-inline-policy-check"

  source {
    owner             = "AWS"
    source_identifier = "IAM_NO_INLINE_POLICY_CHECK"
  }
}

resource "aws_config_config_rule" "iampasswordpolicy" {
  name  = "iam-password-policy"

  source {
    owner             = "AWS"
    source_identifier = "IAM_PASSWORD_POLICY"
  }

  input_parameters = jsonencode({
    MaxPasswordAge = 90,
    MinimumPasswordLength = 14,
    PasswordReusePrevention = 24,
    RequireLowercaseCharacters = true,
    RequireNumbers = true,
    RequireSymbols = true,
    RequireUppercaseCharacters = true
  })
}

resource "aws_config_config_rule" "iampolicyinuse" {
  name  = "iam-policy-in-use"

  source {
    owner             = "AWS"
    source_identifier = "IAM_POLICY_IN_USE"
  }

  input_parameters = jsonencode({
    policyARN = "arn:aws:iam::aws:policy/AWSSupportAccess"
  })
}

resource "aws_config_config_rule" "iampolicynostatementswithadminaccess" {
  name  = "iam-policy-no-statements-with-admin-access"

  source {
    owner             = "AWS"
    source_identifier = "IAM_POLICY_NO_STATEMENTS_WITH_ADMIN_ACCESS"
  }
}

resource "aws_config_config_rule" "iamrootaccesskeycheck" {
  name  = "iam-root-access-key-check"

  source {
    owner             = "AWS"
    source_identifier = "IAM_ROOT_ACCESS_KEY_CHECK"
  }
}

resource "aws_config_config_rule" "iamusergroupmembershipcheck" {
  name  = "iam-user-group-membership-check"

  source {
    owner             = "AWS"
    source_identifier = "IAM_USER_GROUP_MEMBERSHIP_CHECK"
  }
}

resource "aws_config_config_rule" "iamusernopoliciescheck" {
  name  = "iam-user-no-policies-check"

  source {
    owner             = "AWS"
    source_identifier = "IAM_USER_NO_POLICIES_CHECK"
  }
}

resource "aws_config_config_rule" "iamuserunusedcredentialscheck" {
  name  = "iam-user-unused-credentials-check"

  source {
    owner             = "AWS"
    source_identifier = "IAM_USER_UNUSED_CREDENTIALS_CHECK"
  }

  input_parameters = jsonencode({
    maxCredentialUsageAge = 45
  })
}

resource "aws_config_config_rule" "incomingsshdisabled" {
  name  = "restricted-ssh"

  source {
    owner             = "AWS"
    source_identifier = "INCOMING_SSH_DISABLED"
  }
}

resource "aws_config_config_rule" "mfaenabledforiamconsoleaccess" {
  name  = "mfa-enabled-for-iam-console-access"

  source {
    owner             = "AWS"
    source_identifier = "MFA_ENABLED_FOR_IAM_CONSOLE_ACCESS"
  }
}

resource "aws_config_config_rule" "multiregioncloudtrailenabled" {
  name  = "multi-region-cloudtrail-enabled"

  source {
    owner             = "AWS"
    source_identifier = "MULTI_REGION_CLOUD_TRAIL_ENABLED"
  }
}

resource "aws_config_config_rule" "rdssnapshotencrypted" {
  name  = "rds-snapshot-encrypted"

  source {
    owner             = "AWS"
    source_identifier = "RDS_SNAPSHOT_ENCRYPTED"
  }
}

resource "aws_config_config_rule" "rdsstorageencrypted" {
  name  = "rds-storage-encrypted"

  source {
    owner             = "AWS"
    source_identifier = "RDS_STORAGE_ENCRYPTED"
  }
}

resource "aws_config_config_rule" "restrictedincomingtraffic" {
  name  = "restricted-common-ports"

  source {
    owner             = "AWS"
    source_identifier = "RESTRICTED_INCOMING_TRAFFIC"
  }

  input_parameters = jsonencode({
    blockedPort3 = 3389
  })
}

resource "aws_config_config_rule" "rootaccountmfaenabled" {
  name  = "root-account-mfa-enabled"

  source {
    owner             = "AWS"
    source_identifier = "ROOT_ACCOUNT_MFA_ENABLED"
  }
}

resource "aws_config_config_rule" "s3accountlevelpublicaccessblocksperiodic" {
  name  = "s3-account-level-public-access-blocks-periodic"

  source {
    owner             = "AWS"
    source_identifier = "S3_ACCOUNT_LEVEL_PUBLIC_ACCESS_BLOCKS_PERIODIC"
  }

  input_parameters = jsonencode({
    BlockPublicAcls = true,
    BlockPublicPolicy = true,
    IgnorePublicAcls = true,
    RestrictPublicBuckets = true
  })
}

resource "aws_config_config_rule" "s3bucketlevelpublicaccessprohibited" {
  name  = "s3-bucket-level-public-access-prohibited"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_LEVEL_PUBLIC_ACCESS_PROHIBITED"
  }
}

resource "aws_config_config_rule" "s3bucketloggingenabled" {
  name  = "s3-bucket-logging-enabled"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_LOGGING_ENABLED"
  }
}

resource "aws_config_config_rule" "s3bucketpublicreadprohibited" {
  name  = "s3-bucket-public-read-prohibited"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_PUBLIC_READ_PROHIBITED"
  }
}

resource "aws_config_config_rule" "s3bucketpublicwriteprohibited" {
  name  = "s3-bucket-public-write-prohibited"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_PUBLIC_WRITE_PROHIBITED"
  }
}

resource "aws_config_config_rule" "s3bucketversioningenabled" {
  name  = "s3-bucket-versioning-enabled"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  input_parameters = jsonencode({
    isMfaDeleteEnabled = true
  })
}

resource "aws_config_config_rule" "accountcontactdetailsconfigured" {
  name  = "account-contact-details-configured"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "accountsecuritycontactconfigured" {
  name  = "account-security-contact-configured"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "rootaccountregularuse" {
  name  = "root-account-regular-use"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "iamuserconsoleandapiaccessatcreation" {
  name  = "iam-user-console-and-api-access-at-creation"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "iamusersingleaccesskey" {
  name  = "iam-user-single-access-key"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "iamexpiredcertificates" {
  name  = "iam-expired-certificates"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "iamaccessanalyzerenabled" {
  name  = "iam-access-analyzer-enabled"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmunauthorizedapicalls" {
  name  = "alarm-unauthorized-api-calls"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmsigninwithoutmfa" {
  name  = "alarm-sign-in-without-mfa"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmrootaccountuse" {
  name  = "alarm-root-account-use"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmiampolicychange" {
  name  = "alarm-iam-policy-change"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmcloudtrailconfigchange" {
  name  = "alarm-cloudtrail-config-change"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarms3bucketpolicychange" {
  name  = "alarm-s3-bucket-policy-change"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmvpcnetworkgatewaychange" {
  name  = "alarm-vpc-network-gateway-change"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmvpcroutetablechange" {
  name  = "alarm-vpc-route-table-change"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmvpcchange" {
  name  = "alarm-vpc-change"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "alarmorganizationschange" {
  name  = "alarm-organizations-change"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

resource "aws_config_config_rule" "vpcnetworkaclopenadminports" {
  name  = "vpc-networkacl-open-admin-ports"

  source {
    owner             = "AWS"
    source_identifier = "AWS_CONFIG_PROCESS_CHECK"
  }
}

