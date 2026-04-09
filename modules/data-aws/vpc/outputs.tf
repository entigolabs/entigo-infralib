output "vpc_id" {
  description = "The ID of the VPC"
  value       = data.aws_vpc.this.id
}

output "vpc_arn" {
  description = "The ARN of the VPC"
  value       = data.aws_vpc.this.arn
}

output "vpc_cidr_block" {
  description = "The CIDR block of the VPC"
  value       = data.aws_vpc.this.cidr_block
}

output "name" {
  description = "The name of the VPC"
  value       = lookup(data.aws_vpc.this.tags, "Name", "")
}

output "public_subnets" {
  description = "List of IDs of public subnets"
  value       = data.aws_subnets.public.ids
}

output "private_subnets" {
  description = "List of IDs of private subnets"
  value       = data.aws_subnets.private.ids
}

output "control_subnets" {
  description = "List of IDs of control subnets"
  value       = data.aws_subnets.private.ids
}

output "service_subnets" {
  description = "List of IDs of service subnets"
  value       = data.aws_subnets.private.ids
}

output "compute_subnets" {
  description = "List of IDs of compute subnets"
  value       = data.aws_subnets.private.ids
}

output "database_subnets" {
  description = "List of IDs of database subnets"
  value       = []
}

output "elasticache_subnets" {
  description = "List of IDs of elasticache subnets"
  value       = []
}

output "database_subnet_group_name" {
  description = "Name of database subnet group"
  value       = ""
}

output "elasticache_subnet_group_name" {
  description = "Name of elasticache subnet group"
  value       = ""
}

output "default_security_group_id" {
  description = "The ID of the security group created by default on VPC creation"
  value       = data.aws_vpc.this.default_security_group_id
}
