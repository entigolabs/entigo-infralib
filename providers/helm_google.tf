data "google_client_config" "this" {}

data "google_container_cluster" "this" {
  name     = local.hname
  location = data.google_client_config.this.region
}

provider "helm" {
  kubernetes {
    host                   = "https://${data.google_container_cluster.this.endpoint}"
    cluster_ca_certificate = base64decode(data.google_container_cluster.this.master_auth[0].cluster_ca_certificate)
    token                  = data.google_client_config.this.access_token
  }
}
