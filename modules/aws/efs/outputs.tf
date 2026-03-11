################################################################################
# File System
################################################################################
output "arn" {
  description = "List of Amazon Resource Names of the file systems"
  value       = [for k, v in module.efs : v.arn]
}

output "id" {
  description = "List of file system IDs (e.g., `fs-ccfc0d65`)"
  value       = [for k, v in module.efs : v.id]
}

output "dns_name" {
  description = "List of DNS names for the filesystems"
  value       = [for k, v in module.efs : v.dns_name]
}

output "efs_volumes" {
  description = "List of EFS volume names"
  value       = [for k, v in module.efs : k]
}


################################################################################
# Mount Target(s)
################################################################################
output "mount_targets" {
  description = "List of mount targets created and their attributes"
  value       = [for k, v in module.efs : v.mount_targets]
}

################################################################################
# Security Group
################################################################################
output "security_group_arn" {
  description = "List of security group ARNs"
  value       = [for k, v in module.efs : v.security_group_arn]
}

output "security_group_id" {
  description = "List of security group IDs"
  value       = [for k, v in module.efs : v.security_group_id]
}

################################################################################
# Replication Configuration
################################################################################
output "replication_configuration_destination_file_system_id" {
  description = "List of replica file system IDs"
  value       = [for k, v in module.efs : v.replication_configuration_destination_file_system_id]
}
