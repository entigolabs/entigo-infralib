variable "prefix" {
  type = string
}

variable "vpc_id" {
  type = string
  default = ""
}

variable "exclude_domains" {
  type = list(string)
  default = []
}

variable "share_status" {
  type = string
  default = "SHARED_WITH_ME"
}
