variable "prefix" {
  type = string
}

variable "vpc_id" {
  type        = string
  description = "The ID of the VPC to look up"
}

variable "default_security_group_id" {
  type        = string
  description = "The ID of the default security group for the VPC"
  default     = ""
}

variable "pipeline_security_group_id" {
  type        = string
  description = "The ID of the security group used by the infralib pipeline (CodeBuild) for VPC access"
  default     = ""
}
