locals {
  ami_release_version = {
     "AL2023_ARM_64_NVIDIA"       = "1.35.4-20260512"
     "AL2023_ARM_64_STANDARD"     = "1.35.4-20260512"
     "AL2023_x86_64_NEURON"       = "1.35.4-20260512"
     "AL2023_x86_64_NVIDIA"       = "1.35.4-20260512"
     "AL2023_x86_64_STANDARD"     = "1.35.4-20260512"
     "BOTTLEROCKET_ARM_64"        = "1.61.0-8ef015e0"
     "BOTTLEROCKET_ARM_64_FIPS"   = "1.61.0-8ef015e0"
     "BOTTLEROCKET_ARM_64_NVIDIA" = "1.61.0-8ef015e0"
     "BOTTLEROCKET_x86_64"        = "1.61.0-8ef015e0"
     "BOTTLEROCKET_x86_64_FIPS"   = "1.61.0-8ef015e0"
     "BOTTLEROCKET_x86_64_NVIDIA" = "1.61.0-8ef015e0"
  }

  # Node conditions and reasons that trigger EKS node auto repair (non GPU), https://docs.aws.amazon.com/eks/latest/userguide/node-health-nma.html
  node_repair_reasons = {
    Ready                 = ["NodeStatusUnknown", "KubeletNotReady"]
    ContainerRuntimeReady = ["PodStuckTerminating"]
    KernelReady           = ["ForkFailedOutOfPIDs"]
    NetworkingReady       = ["IPAMDNotReady", "IPAMDNotRunning", "InterfaceNotRunning", "InterfaceNotUp", "MissingLoopbackInterface"]
  }

  # Every reason above gets the same wait time
  node_repair_default_overrides = tolist(flatten([
    for condition, reasons in local.node_repair_reasons : [
      for reason in reasons : {
        node_monitoring_condition = condition
        node_unhealthy_reason     = reason
        min_repair_wait_time_mins = var.node_repair_wait_time_mins
        repair_action             = "Replace"
      }
    ]
  ]))

  node_repair_enabled          = try(var.node_repair_config.enabled, false)
  node_repair_custom_overrides = try(var.node_repair_config.node_repair_config_overrides, null)

  # Custom overrides win, otherwise use the defaults
  node_repair_overrides = local.node_repair_custom_overrides != null ? local.node_repair_custom_overrides : local.node_repair_default_overrides

  # No node_repair_config means repair stays off. When repair is enabled add the overrides
  node_repair_config = var.node_repair_config == null ? null : merge(var.node_repair_config, {
    node_repair_config_overrides = local.node_repair_enabled ? local.node_repair_overrides : var.node_repair_config.node_repair_config_overrides
  })
}


#https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest
module "eks-managed-node-group" {
  source  = "terraform-aws-modules/eks/aws//modules/eks-managed-node-group"
  version = "21.26.0"
  use_name_prefix = true
  name                    = substr(var.prefix, 0, 35)
  iam_role_use_name_prefix = true
  iam_role_name           = substr(var.prefix, 0, 35)
  iam_role_additional_policies = zipmap(compact(var.iam_role_additional_policies), compact(var.iam_role_additional_policies))
  launch_template_name    = substr(var.prefix, 0, 35)
  cluster_name            = var.cluster_name
  kubernetes_version         = var.cluster_version
  subnet_ids              = var.subnets
  cluster_primary_security_group_id = var.cluster_primary_security_group_id
  cluster_service_cidr    = var.cluster_service_cidr
  vpc_security_group_ids = var.security_group_ids

  key_name = var.key_name
  pre_bootstrap_user_data = var.pre_bootstrap_user_data
  
  # Conditional cloudinit_pre_nodeadm based on content being non-empty
  cloudinit_pre_nodeadm = var.cloudinit_pre_nodeadm_content != "" ? [
    {
      content_type = "text/x-shellscript"
      content = var.cloudinit_pre_nodeadm_content
    }
  ] : []

  min_size     = var.min_size
  max_size     = var.max_size
  desired_size = var.desired_size

  instance_types = var.instance_types
  capacity_type  = var.capacity_type
  ami_type       = var.ami_type
  ami_release_version = try(coalesce(var.ami_release_version, local.ami_release_version[var.ami_type]), null)
  use_latest_ami_release_version = var.use_latest_ami_release_version
  
  block_device_mappings = var.block_device_mappings != null ? var.block_device_mappings : {
    xvda = {
      device_name = "/dev/xvda"
      ebs = {
        volume_size           = var.volume_size
        volume_iops           = var.volume_iops
        volume_type           = var.volume_type
        encrypted             = true
        kms_key_id            = var.encryption_kms_key_arn != "" ? var.encryption_kms_key_arn : null
        delete_on_termination = true
      }
    }
  }
  
  labels = var.labels
  taints = var.taints
  node_repair_config = local.node_repair_config
  launch_template_tags = merge({
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  }, var.tags)
  tags = merge({
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  }, var.tags)
}
