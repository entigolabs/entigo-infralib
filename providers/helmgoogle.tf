data "google_client_config" "this" {}

data "google_container_cluster" "this" {
  name     = var.gke_cluster_name
  location = data.google_client_config.this.region
}

variable "gke_cluster_name" {
  type = string
}

provider "helmgoogle" {
  kubernetes {
    host                   = "https://${data.google_container_cluster.this.endpoint}"
    cluster_ca_certificate = base64decode(data.google_container_cluster.this.master_auth[0].cluster_ca_certificate)
    token                  = data.google_client_config.this.access_token
  }
}
