variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the identity domain."
  type        = string
}

variable "region" {
  description = "Home region for the domain, e.g. eu-frankfurt-1. Wired from the agent by agent_input.yaml."
  type        = string
}

variable "name_salt" {
  description = "Append a random suffix to the domain name, so a rebuild cannot collide with a torn-down domain still holding its name for a retention period. Leave false for a stable, readable name."
  type        = bool
  default     = false
}
