
output "role_name" {
  description = "The Name of the Role to use with Karpenter Nodes"
  value = aws_iam_role.node.name
}

output "role_arn" {
  description = "The ARN of the Role to use with Karpenter Nodes"
  value = aws_iam_role.node.arn
}
