variable "prefix" {
  type = string
}

variable "service_agent_emails" {
  type    = map(string)
  default = {}
}

variable "labels" {
  type    = map(string)
  default = {}
}

variable "location" {
  type    = string
  default = ""
}

variable "key_ring_name" {
  type    = string
  default = ""
}

variable "create_key_ring" {
  type    = bool
  default = true
}

variable "key_rotation_period" {
  type    = string
  default = null
}

variable "key_destroy_scheduled_duration" {
  type    = string
  default = null
}

variable "key_purpose" {
  type    = string
  default = "ENCRYPT_DECRYPT"
}

variable "key_algorithm" {
  type    = string
  default = "GOOGLE_SYMMETRIC_ENCRYPTION"
}

variable "key_protection_level" {
  type    = string
  default = "SOFTWARE"
}

# Data key
variable "data_key_additional_encrypters" {
  type    = list(string)
  default = []
}

variable "data_key_additional_decrypters" {
  type    = list(string)
  default = []
}

variable "data_key_additional_encrypters_decrypters" {
  type    = list(string)
  default = []
}

# Config key
variable "config_key_additional_encrypters" {
  type    = list(string)
  default = []
}

variable "config_key_additional_decrypters" {
  type    = list(string)
  default = []
}

variable "config_key_additional_encrypters_decrypters" {
  type    = list(string)
  default = []
}

# Telemetry key
variable "telemetry_key_additional_encrypters" {
  type    = list(string)
  default = []
}

variable "telemetry_key_additional_decrypters" {
  type    = list(string)
  default = []
}

variable "telemetry_key_additional_encrypters_decrypters" {
  type    = list(string)
  default = []
}
