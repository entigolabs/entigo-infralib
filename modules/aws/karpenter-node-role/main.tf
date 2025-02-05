resource "aws_iam_role" "node" {
  name        = var.prefix
  description = "${var.prefix} role for Karpeneter nodes"

  assume_role_policy    = data.aws_iam_policy_document.node_assume_role.json
  force_detach_policies = true

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

resource "aws_iam_role_policy_attachment" "node" {
  for_each = { for k, v in merge(
    {
      AmazonEKSWorkerNodePolicy          = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSWorkerNodePolicy"
      AmazonEC2ContainerRegistryReadOnly = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
      AmazonEKS_CNI_Policy               = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKS_CNI_Policy"
      AmazonSSMManagedInstanceCore       = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
    }
  ) : k => v }

  policy_arn = each.value
  role       = aws_iam_role.node.name
}

resource "aws_eks_access_entry" "node" {
  cluster_name  = var.cluster_name
  principal_arn = aws_iam_role.node.arn
  type          = var.access_entry_type
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}
