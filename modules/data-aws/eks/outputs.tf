output "cluster_arn" {
  description = "The Amazon Resource Name (ARN) of the cluster"
  value       = data.aws_eks_cluster.this.arn
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = data.aws_eks_cluster.this.certificate_authority[0].data
}

output "cluster_endpoint" {
  description = "Endpoint for your Kubernetes API server"
  value       = data.aws_eks_cluster.this.endpoint
}

output "cluster_id" {
  description = "The ID of the EKS cluster. Note: currently a value is returned only for local EKS clusters created on Outposts"
  value       = data.aws_eks_cluster.this.id
}

output "cluster_name" {
  description = "The name of the EKS cluster"
  value       = data.aws_eks_cluster.this.name
}

output "cluster_oidc_issuer_url" {
  description = "The URL on the EKS cluster for the OpenID Connect identity provider"
  value       = data.aws_eks_cluster.this.identity[0].oidc[0].issuer
}

output "cluster_version" {
  description = "The Kubernetes version for the cluster"
  value       = data.aws_eks_cluster.this.version
}

output "cluster_platform_version" {
  description = "Platform version for the cluster"
  value       = data.aws_eks_cluster.this.platform_version
}

output "cluster_status" {
  description = "Status of the EKS cluster. One of `CREATING`, `ACTIVE`, `DELETING`, `FAILED`"
  value       = data.aws_eks_cluster.this.status
}

output "cluster_primary_security_group_id" {
  description = "Cluster security group that was created by Amazon EKS for the cluster. Managed node groups use this security group for control-plane-to-data-plane communication. Referred to as 'Cluster security group' in the EKS console"
  value       = data.aws_eks_cluster.this.vpc_config[0].cluster_security_group_id
}

output "kms_key_arn" {
  description = "The Amazon Resource Name (ARN) of the key"
  value       = ""
}

output "kms_key_id" {
  description = "The globally unique identifier for the key"
  value       = ""
}

output "kms_key_policy" {
  description = "The IAM resource policy set on the key"
  value       = ""
}

output "cluster_security_group_arn" {
  description = "Amazon Resource Name (ARN) of the cluster security group"
  value       = length(data.aws_security_group.cluster) > 0 ? data.aws_security_group.cluster[0].arn : ""
}

output "cluster_security_group_id" {
  description = "ID of the cluster security group"
  value       = length(data.aws_security_group.cluster) > 0 ? data.aws_security_group.cluster[0].id : ""
}

output "cluster_service_cidr" {
  description = "The CIDR block where Kubernetes pod and service IP addresses are assigned from"
  value       = data.aws_eks_cluster.this.kubernetes_network_config[0].service_ipv4_cidr
}

output "node_security_group_arn" {
  description = "Amazon Resource Name (ARN) of the node shared security group"
  value       = length(data.aws_security_group.node) > 0 ? data.aws_security_group.node[0].arn : ""
}

output "node_security_group_id" {
  description = "ID of the node shared security group"
  value       = length(data.aws_security_group.node) > 0 ? data.aws_security_group.node[0].id : ""
}

output "oidc_provider" {
  description = "The OpenID Connect identity provider (issuer URL without leading `https://`)"
  value       = replace(data.aws_eks_cluster.this.identity[0].oidc[0].issuer, "https://", "")
}

output "oidc_provider_arn" {
  description = "The ARN of the OIDC Provider if `enable_irsa = true`"
  value       = data.aws_iam_openid_connect_provider.this.arn
}

output "cluster_tls_certificate_sha1_fingerprint" {
  description = "The SHA1 fingerprint of the public key of the cluster's certificate"
  value       = data.tls_certificate.this.certificates[0].sha1_fingerprint
}

output "cluster_iam_role_name" {
  description = "IAM role name of the EKS cluster"
  value       = data.aws_iam_role.cluster.name
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN of the EKS cluster"
  value       = data.aws_eks_cluster.this.role_arn
}

output "cluster_iam_role_unique_id" {
  description = "Stable and unique string identifying the IAM role"
  value       = data.aws_iam_role.cluster.unique_id
}

output "cluster_addons" {
  description = "Map of attribute maps for all EKS cluster addons enabled"
  value = merge(
    {
      coredns              = data.aws_eks_addon.coredns
      "kube-proxy"         = data.aws_eks_addon.kube_proxy
      "vpc-cni"            = data.aws_eks_addon.vpc_cni
      "aws-ebs-csi-driver" = data.aws_eks_addon.ebs_csi
    },
    var.enable_efs_csi ? { "aws-efs-csi-driver" = data.aws_eks_addon.efs_csi[0] } : {}
  )
}

output "efs_csi_service_account_role_arn" {
  description = "AWS EKS EFS CSI Service Account Role ARN"
  value       = var.enable_efs_csi ? data.aws_eks_addon.efs_csi[0].service_account_role_arn : ""
}

output "cluster_identity_providers" {
  description = "Map of attribute maps for all EKS identity providers enabled"
  value       = {}
}

output "cloudwatch_log_group_name" {
  description = "Name of cloudwatch log group created"
  value       = data.aws_cloudwatch_log_group.this.name
}

output "cloudwatch_log_group_arn" {
  description = "Arn of cloudwatch log group created"
  value       = data.aws_cloudwatch_log_group.this.arn
}

output "fargate_profiles" {
  description = "Map of attribute maps for all EKS Fargate Profiles created"
  value       = {}
}

output "eks_managed_node_groups" {
  description = "Map of attribute maps for all EKS managed node groups created"
  value = {
    for name, ng in data.aws_eks_node_group.all : name => {
      node_group_arn       = ng.arn
      node_group_name      = ng.node_group_name
      node_group_status    = ng.status
      node_group_resources = ng.resources
      iam_role_arn         = ng.node_role_arn
      iam_role_name        = element(split("/", ng.node_role_arn), length(split("/", ng.node_role_arn)) - 1)
    }
  }
}

output "eks_managed_node_groups_autoscaling_group_names" {
  description = "List of the autoscaling group names created by EKS managed node groups"
  value = flatten([
    for name, ng in data.aws_eks_node_group.all : [
      for asg in ng.resources[0].autoscaling_groups : asg.name
    ]
  ])
}

output "self_managed_node_groups" {
  description = "Map of attribute maps for all self managed node groups created"
  value       = {}
}

output "self_managed_node_groups_autoscaling_group_names" {
  description = "List of the autoscaling group names created by self-managed node groups"
  value       = []
}

output "account" {
  description = "Cluster account"
  value       = data.aws_caller_identity.current.account_id
}

output "region" {
  description = "Cluster region"
  value       = data.aws_region.current.id
}
