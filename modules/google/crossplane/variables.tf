variable "prefix" {
  type = string
}

variable "ksa_name" {
  type = string
  description = "Kubernetes service account name for crossplane"
  default = "crossplane"
}

variable "kns_name" {
  type = string
  description = "Kubernetes namespace name for crossplane"
  default = "crossplane"
}

variable "project_number" {
  type = string
  description = "Project number"
}

data "google_client_config" "this" {}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
  member = "principal://iam.googleapis.com/projects/${var.project_number}/locations/global/workloadIdentityPools/${data.google_client_config.this.project}.svc.id.goog/subject/ns/${var.kns_name}/sa/${var.ksa_name}"
}
