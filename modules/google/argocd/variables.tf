variable "prefix" {
  type = string
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}