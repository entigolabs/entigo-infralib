 
resource "google_compute_network" "vpc" {
  name                                      = local.hname
}


resource "google_secret_manager_secret" "vpc_id" {
  secret_id = "entigo-infralib-${local.hname}-vpc_id"

  annotations = {
    product = "entigo-infralib"
    hname = local.hname
    workspace = terraform.workspace
    prefix = var.prefix
    parameter = "vpc_id"
  }
  
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "vpc_id" {
  secret = google_secret_manager_secret.vpc_id.id
  secret_data = google_compute_network.vpc.id
}
