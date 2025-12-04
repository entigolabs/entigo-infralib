variable "prefix" {
  type = string
}

variable "cluster_name" {
  type  = string
}

variable "cluster_region" {
  type  = string
}

variable "min_size" {
  type     = number
  nullable = false
  default  = 1
}

variable "max_size" {
  type     = number
  nullable = false
  default  = 3
}

variable "instance_type" {
  type    = string
  default = "e2-standard-2"
}

variable "node_locations" {
  type    = list(string)
  default = []
}

variable "location_policy" {
  type    = string
  default = "BALANCED"
}

variable "spot_nodes" {
  type    = bool
  default = false
}

variable "volume_size" {
  type    = number
  default = 50
}

variable "max_pods" {
  type    = number
  default = 64
}

variable "volume_type" {
  type    = string
  default = "pd-standard"
}

variable "service_account" {
  type        = string
}
