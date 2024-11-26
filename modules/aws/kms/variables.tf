variable "prefix" {
  type = string
}


#current modues:
# disabled - disable encryption
# kms - create a AWS managed KMS key (recommended)
variable "mode" {
  type     = string
  nullable = false
  default  = "kms"
}

variable "deletion_window_in_days" {
  type     = number
  nullable = true
  default  = 30
}

variable "enable_key_rotation" {
  type     = bool
  nullable = true
  default  = false
}

variable "multi_region" {
  type     = bool
  nullable = true
  default  = true
}
