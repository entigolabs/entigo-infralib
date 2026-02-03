variable "prefix" {
  type = string
}

variable "additional_services" {
  type    = list(string)
  default = []
}

variable "additional_services_without_identity" {
  type    = list(string)
  default = []
}

