variable "prefix" {
  type = string
}

variable "zone" {
    type = string
}

variable "domain" {
    type = string
}

variable "type" {
    type = string
}

variable "ksa_name" {
  type = string
  description = "Kubernetes service account name for external-dns"
  default = "external-dns"
}

variable "kns_name" {
  type = string
  description = "Kubernetes namespace name for external-dns"
  default = "external-dns"
}

variable "project_number" {
  type = string
  description = "Project number"
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
  member = "principal://iam.googleapis.com/projects/${var.project_number}/locations/global/workloadIdentityPools/${data.google_client_config.this.project}.svc.id.goog/subject/ns/${var.kns_name}/sa/${var.ksa_name}"
}
