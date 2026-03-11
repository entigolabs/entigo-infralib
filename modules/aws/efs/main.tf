#https://registry.terraform.io/modules/terraform-aws-modules/efs/aws/latest
module "efs" {
  source = "terraform-aws-modules/efs/aws"
  version = "2.2.0"

  for_each = toset(var.efs_volumes)

  name           = "${var.prefix}-${each.key}"
  encrypted      = var.encrypted
  kms_key_arn = var.kms_key_arn != "" ? var.kms_key_arn : null
  attach_policy  = var.attach_policy

  # Merge built-in statements with user-provided ones (user can override built-in keys)
  policy_statements = merge(
    # Only create allow_mount_write when efs_csi_service_account_role_arn is set
    var.efs_csi_service_account_role_arn != "" ? {
      allow_mount_write = {
        effect  = "Allow"
        actions = ["elasticfilesystem:ClientMount", "elasticfilesystem:ClientWrite"]
        principals = [
          {
            type        = "AWS"
            identifiers = [var.efs_csi_service_account_role_arn]
          }
        ]
      }
    } : {},
    var.policy_statements != null ? var.policy_statements : {}
  )

  # Dynamically generate mount targets from subnet ID list
  mount_targets = {
    for subnet_id in var.mount_targets_subnet_id : subnet_id => {
      subnet_id = subnet_id
    }
  }

  security_group_description = "${var.prefix}-${each.key} EFS security group"
  security_group_vpc_id      = var.security_group_vpc_id
  create_security_group      = var.create_security_group


  security_group_use_name_prefix = var.security_group_use_name_prefix

  # Generate per-CIDR ingress rules and merge with user-provided ones
  security_group_ingress_rules = merge(
    {
      for idx, cidr in var.security_group_cidrs : "nfs_${idx}" => {
        description = "NFS ingress CIDRs"
        cidr_ipv4   = cidr
      }
    },
    var.security_group_ingress_rules
  )

  security_group_egress_rules = var.security_group_egress_rules

  performance_mode = var.performance_mode
  lifecycle_policy = var.lifecycle_policy
  provisioned_throughput_in_mibps = var.provisioned_throughput_in_mibps
  throughput_mode = var.throughput_mode

  create_backup_policy = var.create_backup_policy
  enable_backup_policy = var.enable_backup_policy
  bypass_policy_lockout_safety_check = var.bypass_policy_lockout_safety_check
  deny_nonsecure_transport = var.deny_nonsecure_transport
  deny_nonsecure_transport_via_mount_target = var.deny_nonsecure_transport_via_mount_target
  create_replication_configuration = var.create_replication_configuration
  replication_configuration_destination = var.replication_configuration_destination

  protection = var.protection

  tags = merge(
        {
          Terraform  = "true"
          Prefix     = var.prefix
          created-by = "entigo-infralib"
        },
        var.tags
      )

}
