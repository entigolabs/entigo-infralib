output "launch_template_id" {
  description = "The ID of the launch template"
  value       = module.eks-managed-node-group.launch_template_id
}

output "launch_template_arn" {
  description = "The ARN of the launch template"
  value       = module.eks-managed-node-group.launch_template_arn
}

output "launch_template_latest_version" {
  description = "The latest version of the launch template"
  value       = module.eks-managed-node-group.launch_template_latest_version
}

output "launch_template_name" {
  description = "The name of the launch template"
  value       = module.eks-managed-node-group.launch_template_name
}

output "node_group_arn" {
  description = "Amazon Resource Name (ARN) of the EKS Node Group"
  value       = module.eks-managed-node-group.node_group_arn
}

output "node_group_id" {
  description = "EKS Cluster name and EKS Node Group name separated by a colon (`:`)"
  value       = module.eks-managed-node-group.node_group_id
}

output "node_group_resources" {
  description = "List of objects containing information about underlying resources"
  value       = module.eks-managed-node-group.node_group_resources
}

output "node_group_autoscaling_group_names" {
  description = "List of the autoscaling group names"
  value       = module.eks-managed-node-group.node_group_autoscaling_group_names
}

output "node_group_status" {
  description = "Status of the EKS Node Group"
  value       = module.eks-managed-node-group.node_group_status
}

output "node_group_labels" {
  description = "Map of labels applied to the node group"
  value       = module.eks-managed-node-group.node_group_labels
}

output "node_group_taints" {
  description = "List of objects containing information about taints applied to the node group"
  value       = module.eks-managed-node-group.node_group_taints
}

output "autoscaling_group_schedule_arns" {
  description = "ARNs of autoscaling group schedules"
  value       = module.eks-managed-node-group.autoscaling_group_schedule_arns
}
