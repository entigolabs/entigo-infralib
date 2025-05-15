locals {
  # https://docs.aws.amazon.com/config/latest/developerguide/resource-config-reference.html
  required_tags_custom_resource_types = [
    "AWS::IAM::User",
    "AWS::IAM::Policy",
    "AWS::IAM::User",
    "AWS::IAM::Policy",
    "AWS::IAM::Role",
    "AWS::Lambda::Function",
    "AWS::SNS::Topic",
    "AWS::SQS::Queue",
    "AWS::EKS::Cluster",
    "AWS::CloudFront::Distribution",
    "AWS::KMS::Key",
    "AWS::ACM::Certificate",
    "AWS::AutoScaling::AutoScalingGroup",
    "AWS::CloudFormation::Stack",
    "AWS::CodeBuild::Project",
    "AWS::DynamoDB::Table",
    "AWS::EC2::CustomerGateway",
    "AWS::EC2::Instance",
    "AWS::EC2::InternetGateway",
    "AWS::EC2::NetworkAcl",
    "AWS::EC2::NetworkInterface",
    "AWS::EC2::RouteTable",
    "AWS::EC2::SecurityGroup",
    "AWS::EC2::Subnet",
    "AWS::EC2::Volume",
    "AWS::EC2::VPC",
    "AWS::EC2::VPNConnection",
    "AWS::EC2::VPNGateway",
    "AWS::ElasticLoadBalancing::LoadBalancer",
    "AWS::ElasticLoadBalancingV2::LoadBalancer",
    "AWS::RDS::DBInstance",
    "AWS::RDS::DBSecurityGroup",
    "AWS::RDS::DBSnapshot",
    "AWS::RDS::DBSubnetGroup",
    "AWS::RDS::EventSubscription",
    "AWS::S3::Bucket"
  ]

  required_tags_custom_resource_types_formatted = join(",\n  ", formatlist("\"%s\"", local.required_tags_custom_resource_types))

  required_tags_custom_rules = join("\n\n", [
    for tag in var.required_tag_keys : <<-EOT
      rule check_required_tag_${replace(tag, "-", "_")} when resourceType in [
        ${local.required_tags_custom_resource_types_formatted}
      ] {
        resourceName == /AWSServiceRoleFor(?i)/ or 
        resourceName == /OrganizationAccountAccessRole(?i)/ or
        resourceName == /AWSReservedSSO_AdministratorAccess(?i)/ or
        resourceName == /AmazonEKS_(?i)/ or
        resourceName == /VPC-CNI-(?i)/ or
        tags["${tag}"] !empty
      }
    EOT
  ])
}

# resource "aws_config_config_rule" "required_tags_custom" {
#   name = "required-tags-custom"
#   scope {
#     compliance_resource_types = local.required_tags_custom_resource_types
#   }
#   source {
#     owner = "CUSTOM_POLICY"
#     source_detail {
#       message_type = "ConfigurationItemChangeNotification"
#     }
#     custom_policy_details {
#       policy_runtime = "guard-2.x.x"
#       policy_text    = local.tag_rules
#     }
#   }
# }

# https://github.com/awslabs/aws-config-rules/tree/master/aws-config-conformance-packs
resource "aws_config_conformance_pack" "operational_best_practices" {
  name  = "${var.prefix}-operational-best-practices"
  count = var.operational_best_practices_conformance_pack_enabled ? 1 : 0

  template_body = <<EOT
Parameters:
  AccessKeysRotatedParamMaxAccessKeyAge:
    Default: '90'
    Type: String
  IamPasswordPolicyParamMaxPasswordAge:
    Default: '90'
    Type: String
  IamPasswordPolicyParamMinimumPasswordLength:
    Default: '14'
    Type: String
  IamPasswordPolicyParamPasswordReusePrevention:
    Default: '24'
    Type: String
  IamPasswordPolicyParamRequireLowercaseCharacters:
    Default: 'true'
    Type: String
  IamPasswordPolicyParamRequireNumbers:
    Default: 'true'
    Type: String
  IamPasswordPolicyParamRequireSymbols:
    Default: 'true'
    Type: String
  IamPasswordPolicyParamRequireUppercaseCharacters:
    Default: 'true'
    Type: String
  IamUserUnusedCredentialsCheckParamMaxCredentialUsageAge:
    Default: '45'
    Type: String
  RestrictedIncomingTrafficParamBlockedPort3:
    Default: '3389'
    Type: String
Resources:
%{if length(var.required_tag_keys) > 0}
  CheckRequiredTags:
    Properties:
      ConfigRuleName: required-tags-custom
      Description: Ensure required tags are present on resources.
      Scope:
        ComplianceResourceTypes: ${jsonencode(local.required_tags_custom_resource_types)}
      Source:
        Owner: CUSTOM_POLICY
        SourceDetails:
          - EventSource: aws.config
            MessageType: ConfigurationItemChangeNotification
        CustomPolicyDetails:
          PolicyRuntime: guard-2.x.x
          PolicyText: |
            ${indent(12, local.required_tags_custom_rules)}
    Type: AWS::Config::ConfigRule
%{endif}
  AccessKeysRotated:
    Properties:
      ConfigRuleName: access-keys-rotated
      InputParameters:
        maxAccessKeyAge:
          Fn::If:
          - accessKeysRotatedParamMaxAccessKeyAge
          - Ref: AccessKeysRotatedParamMaxAccessKeyAge
          - Ref: AWS::NoValue
      Source:
        Owner: AWS
        SourceIdentifier: ACCESS_KEYS_ROTATED
    Type: AWS::Config::ConfigRule
  CloudTrailCloudWatchLogsEnabled:
    Properties:
      ConfigRuleName: cloud-trail-cloud-watch-logs-enabled
      Source:
        Owner: AWS
        SourceIdentifier: CLOUD_TRAIL_CLOUD_WATCH_LOGS_ENABLED
    Type: AWS::Config::ConfigRule
  Ec2EbsEncryptionByDefault:
    Properties:
      ConfigRuleName: ec2-ebs-encryption-by-default
      Source:
        Owner: AWS
        SourceIdentifier: EC2_EBS_ENCRYPTION_BY_DEFAULT
    Type: AWS::Config::ConfigRule
  EncryptedVolumes:
    Properties:
      ConfigRuleName: encrypted-volumes
      Scope:
        ComplianceResourceTypes:
        - AWS::EC2::Volume
      Source:
        Owner: AWS
        SourceIdentifier: ENCRYPTED_VOLUMES
    Type: AWS::Config::ConfigRule
  IamNoInlinePolicyCheck:
    Properties:
      ConfigRuleName: iam-no-inline-policy-check
      Scope:
        ComplianceResourceTypes:
        - AWS::IAM::User
        - AWS::IAM::Role
        - AWS::IAM::Group
      Source:
        Owner: AWS
        SourceIdentifier: IAM_NO_INLINE_POLICY_CHECK
    Type: AWS::Config::ConfigRule
  IamPasswordPolicy:
    Properties:
      ConfigRuleName: iam-password-policy
      InputParameters:
        MaxPasswordAge:
          Fn::If:
          - iamPasswordPolicyParamMaxPasswordAge
          - Ref: IamPasswordPolicyParamMaxPasswordAge
          - Ref: AWS::NoValue
        MinimumPasswordLength:
          Fn::If:
          - iamPasswordPolicyParamMinimumPasswordLength
          - Ref: IamPasswordPolicyParamMinimumPasswordLength
          - Ref: AWS::NoValue
        PasswordReusePrevention:
          Fn::If:
          - iamPasswordPolicyParamPasswordReusePrevention
          - Ref: IamPasswordPolicyParamPasswordReusePrevention
          - Ref: AWS::NoValue
        RequireLowercaseCharacters:
          Fn::If:
          - iamPasswordPolicyParamRequireLowercaseCharacters
          - Ref: IamPasswordPolicyParamRequireLowercaseCharacters
          - Ref: AWS::NoValue
        RequireNumbers:
          Fn::If:
          - iamPasswordPolicyParamRequireNumbers
          - Ref: IamPasswordPolicyParamRequireNumbers
          - Ref: AWS::NoValue
        RequireSymbols:
          Fn::If:
          - iamPasswordPolicyParamRequireSymbols
          - Ref: IamPasswordPolicyParamRequireSymbols
          - Ref: AWS::NoValue
        RequireUppercaseCharacters:
          Fn::If:
          - iamPasswordPolicyParamRequireUppercaseCharacters
          - Ref: IamPasswordPolicyParamRequireUppercaseCharacters
          - Ref: AWS::NoValue
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
  IamPolicyNoStatementsWithAdminAccess:
    Properties:
      ConfigRuleName: iam-policy-no-statements-with-admin-access
      Scope:
        ComplianceResourceTypes:
        - AWS::IAM::Policy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_POLICY_NO_STATEMENTS_WITH_ADMIN_ACCESS
    Type: AWS::Config::ConfigRule
  IamRootAccessKeyCheck:
    Properties:
      ConfigRuleName: iam-root-access-key-check
      Source:
        Owner: AWS
        SourceIdentifier: IAM_ROOT_ACCESS_KEY_CHECK
    Type: AWS::Config::ConfigRule
  IamUserUnusedCredentialsCheck:
    Properties:
      ConfigRuleName: iam-user-unused-credentials-check
      InputParameters:
        maxCredentialUsageAge:
          Fn::If:
          - iamUserUnusedCredentialsCheckParamMaxCredentialUsageAge
          - Ref: IamUserUnusedCredentialsCheckParamMaxCredentialUsageAge
          - Ref: AWS::NoValue
      Source:
        Owner: AWS
        SourceIdentifier: IAM_USER_UNUSED_CREDENTIALS_CHECK
    Type: AWS::Config::ConfigRule
  IncomingSshDisabled:
    Properties:
      ConfigRuleName: restricted-ssh
      Scope:
        ComplianceResourceTypes:
        - AWS::EC2::SecurityGroup
      Source:
        Owner: AWS
        SourceIdentifier: INCOMING_SSH_DISABLED
    Type: AWS::Config::ConfigRule
  MfaEnabledForIamConsoleAccess:
    Properties:
      ConfigRuleName: mfa-enabled-for-iam-console-access
      Source:
        Owner: AWS
        SourceIdentifier: MFA_ENABLED_FOR_IAM_CONSOLE_ACCESS
    Type: AWS::Config::ConfigRule
  MultiRegionCloudTrailEnabled:
    Properties:
      ConfigRuleName: multi-region-cloudtrail-enabled
      Source:
        Owner: AWS
        SourceIdentifier: MULTI_REGION_CLOUD_TRAIL_ENABLED
    Type: AWS::Config::ConfigRule
  RdsSnapshotEncrypted:
    Properties:
      ConfigRuleName: rds-snapshot-encrypted
      Scope:
        ComplianceResourceTypes:
        - AWS::RDS::DBSnapshot
        - AWS::RDS::DBClusterSnapshot
      Source:
        Owner: AWS
        SourceIdentifier: RDS_SNAPSHOT_ENCRYPTED
    Type: AWS::Config::ConfigRule
  RdsStorageEncrypted:
    Properties:
      ConfigRuleName: rds-storage-encrypted
      Scope:
        ComplianceResourceTypes:
        - AWS::RDS::DBInstance
      Source:
        Owner: AWS
        SourceIdentifier: RDS_STORAGE_ENCRYPTED
    Type: AWS::Config::ConfigRule
  RestrictedIncomingTraffic:
    Properties:
      ConfigRuleName: restricted-common-ports
      InputParameters:
        blockedPort3:
          Fn::If:
          - restrictedIncomingTrafficParamBlockedPort3
          - Ref: RestrictedIncomingTrafficParamBlockedPort3
          - Ref: AWS::NoValue
      Scope:
        ComplianceResourceTypes:
        - AWS::EC2::SecurityGroup
      Source:
        Owner: AWS
        SourceIdentifier: RESTRICTED_INCOMING_TRAFFIC
    Type: AWS::Config::ConfigRule
  RootAccountMfaEnabled:
    Properties:
      ConfigRuleName: root-account-mfa-enabled
      Source:
        Owner: AWS
        SourceIdentifier: ROOT_ACCOUNT_MFA_ENABLED
    Type: AWS::Config::ConfigRule
  AccountContactDetailsConfigured:
    Properties:
      ConfigRuleName: account-contact-details-configured
      Description: Ensure the contact email and telephone number for AWS accounts are current and map to more than one individual in your organization. Within the My Account section of the console ensure correct information is specified in the Contact Information section.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AccountSecurityContactConfigured:
    Properties:
      ConfigRuleName: account-security-contact-configured
      Description: Ensure the contact email and telephone number for the your organizations security team are current. Within the My Account section of the AWS Management Console ensure the correct information is specified in the Security section.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  RootAccountRegularUse:
    Properties:
      ConfigRuleName: root-account-regular-use
      Description: Ensure the use of the root account is avoided for everyday tasks. Within IAM, run a credential report to examine when the root user was last used.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  IAMUserConsoleAndAPIAccessAtCreation:
    Properties:
      ConfigRuleName: iam-user-console-and-api-access-at-creation
      Description: Ensure access keys are not setup during the initial user setup for all IAM users that have a console password. For all IAM users with console access, compare the user 'Creation time` to the Access Key `Created` date.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  IAMUserSingleAccessKey:
    Properties:
      ConfigRuleName: iam-user-single-access-key
      Description: Ensure there is only one active access key available for any single IAM user. For all IAM users check that there is only one active key used within the Security Credentials tab for each user within IAM.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  IAMExpiredCertificates:
    Properties:
      ConfigRuleName: iam-expired-certificates
      Description: Ensure that all the expired SSL/TLS certificates stored in IAM are removed. From the command line with the installed AWS CLI run the 'aws iam list-server-certificates' command and determine if there are any expired server certificates.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  IAMAccessAnalyzerEnabled:
    Properties:
      ConfigRuleName: iam-access-analyzer-enabled
      Description: Ensure that IAM Access analyzer is enabled. Within the IAM section of the console, select Access analyzer and ensure that the STATUS is set to Active.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmUnauthorizedAPIcalls:
    Properties:
      ConfigRuleName: alarm-unauthorized-api-calls
      Description: Ensure a log metric filter and an alarm exists for unauthorized API calls.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmSignInWithoutMFA:
    Properties:
      ConfigRuleName: alarm-sign-in-without-mfa
      Description: Ensure a log metric filter and an alarm exists for AWS Management Console sign-in without Multi-Factor Authentication (MFA).
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmRootAccountUse:
    Properties:
      ConfigRuleName: alarm-root-account-use
      Description: Ensure a log metric filter and an alarm exists for usage of the root account.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmIAMpolicyChange:
    Properties:
      ConfigRuleName: alarm-iam-policy-change
      Description: Ensure a log metric filter and an alarm exists for IAM policy changes.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmCloudtrailConfigChange:
    Properties:
      ConfigRuleName: alarm-cloudtrail-config-change
      Description: Ensure a log metric filter and an alarm exists for AWS CloudTrail configuration changes.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmVPCNetworkGatewayChange:
    Properties:
      ConfigRuleName: alarm-vpc-network-gateway-change
      Description: Ensure a log metric filter and an alarm exists for changes to network gateways.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmVPCroutetableChange:
    Properties:
      ConfigRuleName: alarm-vpc-route-table-change
      Description: Ensure a log metric filter and an alarm exists for route table changes.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmVPCChange:
    Properties:
      ConfigRuleName: alarm-vpc-change
      Description: Ensure a log metric filter and an alarm exists for Amazon Virtual Private Cloud (VPC) changes.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  AlarmOrganizationsChange:
    Properties:
      ConfigRuleName: alarm-organizations-change
      Description: Ensure a log metric filter and an alarm exists for AWS Organizations changes.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
  VPCNetworkACLOpenAdminPorts:
    Properties:
      ConfigRuleName: vpc-networkacl-open-admin-ports
      Description: Ensure no network ACLs allow public ingress to the remote server administration ports. Within the VPC section of the console, ensure there are network ACLs with a source of '0.0.0.0/0' with allowing ports or port ranges including remote server admin ports.
      Source:
        Owner: AWS
        SourceIdentifier: AWS_CONFIG_PROCESS_CHECK
    Type: AWS::Config::ConfigRule
Conditions:
  accessKeysRotatedParamMaxAccessKeyAge:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: AccessKeysRotatedParamMaxAccessKeyAge
  iamPasswordPolicyParamMaxPasswordAge:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: IamPasswordPolicyParamMaxPasswordAge
  iamPasswordPolicyParamMinimumPasswordLength:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: IamPasswordPolicyParamMinimumPasswordLength
  iamPasswordPolicyParamPasswordReusePrevention:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: IamPasswordPolicyParamPasswordReusePrevention
  iamPasswordPolicyParamRequireLowercaseCharacters:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: IamPasswordPolicyParamRequireLowercaseCharacters
  iamPasswordPolicyParamRequireNumbers:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: IamPasswordPolicyParamRequireNumbers
  iamPasswordPolicyParamRequireSymbols:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: IamPasswordPolicyParamRequireSymbols
  iamPasswordPolicyParamRequireUppercaseCharacters:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: IamPasswordPolicyParamRequireUppercaseCharacters
  iamUserUnusedCredentialsCheckParamMaxCredentialUsageAge:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: IamUserUnusedCredentialsCheckParamMaxCredentialUsageAge
  restrictedIncomingTrafficParamBlockedPort3:
    Fn::Not:
    - Fn::Equals:
      - ''
      - Ref: RestrictedIncomingTrafficParamBlockedPort3
EOT

  depends_on = [aws_config_configuration_recorder.config_rules]
}
