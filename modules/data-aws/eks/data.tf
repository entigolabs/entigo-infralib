data "aws_eks_cluster" "this" {
  name = var.cluster_name
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

data "aws_iam_openid_connect_provider" "this" {
  url = data.aws_eks_cluster.this.identity[0].oidc[0].issuer
}

data "tls_certificate" "this" {
  url = data.aws_eks_cluster.this.identity[0].oidc[0].issuer
}

data "aws_iam_role" "cluster" {
  name = element(split("/", data.aws_eks_cluster.this.role_arn), length(split("/", data.aws_eks_cluster.this.role_arn)) - 1)
}

data "aws_security_group" "cluster" {
  count = length(tolist(data.aws_eks_cluster.this.vpc_config[0].security_group_ids)) > 0 ? 1 : 0
  id    = tolist(data.aws_eks_cluster.this.vpc_config[0].security_group_ids)[0]
}

data "aws_security_groups" "node" {
  count = var.node_security_group_id == null ? 1 : 0
  filter {
    name   = "tag:karpenter.sh/discovery"
    values = [var.cluster_name]
  }
}

locals {
  node_sg_ids = var.node_security_group_id != null ? [var.node_security_group_id] : (length(data.aws_security_groups.node) > 0 ? data.aws_security_groups.node[0].ids : [])
}

data "aws_security_group" "node" {
  count = length(local.node_sg_ids) > 0 ? 1 : 0
  id    = length(local.node_sg_ids) > 0 ? local.node_sg_ids[0] : null
}

data "aws_cloudwatch_log_group" "this" {
  name = "/aws/eks/${var.cluster_name}/cluster"
}

data "aws_eks_node_groups" "all" {
  cluster_name = var.cluster_name
}

data "aws_eks_node_group" "all" {
  for_each        = data.aws_eks_node_groups.all.names
  cluster_name    = var.cluster_name
  node_group_name = each.value
}

data "aws_eks_addon" "coredns" {
  cluster_name = var.cluster_name
  addon_name   = "coredns"
}

data "aws_eks_addon" "kube_proxy" {
  cluster_name = var.cluster_name
  addon_name   = "kube-proxy"
}

data "aws_eks_addon" "vpc_cni" {
  cluster_name = var.cluster_name
  addon_name   = "vpc-cni"
}

data "aws_eks_addon" "ebs_csi" {
  cluster_name = var.cluster_name
  addon_name   = "aws-ebs-csi-driver"
}

data "aws_eks_addon" "efs_csi" {
  count        = var.enable_efs_csi ? 1 : 0
  cluster_name = var.cluster_name
  addon_name   = "aws-efs-csi-driver"
}
