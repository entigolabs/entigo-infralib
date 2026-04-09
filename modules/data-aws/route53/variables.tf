variable "prefix" {
  type = string
}

variable "zone_name" {
  type        = string
  description = "The name of the Route53 hosted zone to look up"
}

variable "private_zone" {
  type        = bool
  default     = false
  description = "Whether the zone is private"
}
