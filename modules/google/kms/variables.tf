variable "prefix" {
  type = string
}

variable "kms_key_ring_name" {
  type = string
  default = ""
}

variable "create_kms_key_ring" {
  type = bool
  default = true
}

variable "kms_key_rotation_period" {
  type = string
  default = null
}

variable "kms_destroy_scheduled_duration" {
  type = string
  default = null
}

variable "kms_data_key_encrypters" {
  type = list(string)
  default = []
}

variable "kms_data_key_decrypters" {
  type = list(string)
  default = []
}

variable "kms_config_key_encrypters" {
  type = list(string)
  default = []
}

variable "kms_config_key_decrypters" {
  type = list(string)
  default = []
}

variable "kms_telemetry_key_encrypters" {
  type = list(string)
  default = []
}

variable "kms_telemetry_key_decrypters" {
  type = list(string)
  default = []
}

variable "labels" {
  type = map(string)
  default = {}
}