#https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest
module "eks-managed-node-group" {
  source  = "terraform-aws-modules/eks/aws//modules/eks-managed-node-group"
  version = "20.31.6"
  use_name_prefix = true
  name                    = substr(var.prefix, 0, 35)
  iam_role_use_name_prefix = true
  iam_role_name           = substr(var.prefix, 0, 35)
  launch_template_name    = substr(var.prefix, 0, 35)
  cluster_name            = var.cluster_name
  cluster_version         = var.cluster_version
  subnet_ids              = var.subnets
  cluster_primary_security_group_id = var.cluster_primary_security_group_id
  cluster_service_cidr    = var.cluster_service_cidr
  vpc_security_group_ids            = [var.node_security_group_id]
  
  pre_bootstrap_user_data = var.pre_bootstrap_user_data
  remote_access = var.remote_access
  
  min_size     = var.min_size
  max_size     = var.max_size
  desired_size = var.desired_size

  instance_types = var.instance_types
  capacity_type  = var.capacity_type
  ami_type       = var.ami_type
  
  block_device_mappings = {
    xvda = {
      device_name = "/dev/xvda"
      ebs = {
        volume_size           = var.volume_size
        volume_iops           = var.volume_iops
        volume_type           = var.volume_type
        encrypted             = var.encryption_kms_key_arn != "" ? true : false
        kms_key_id            = var.encryption_kms_key_arn != "" ? var.encryption_kms_key_arn : null
        delete_on_termination = true
      }
    }
  }
  
  labels = var.labels
  taints = var.taints
  launch_template_tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}
