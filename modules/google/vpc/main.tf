 
 resource "google_compute_network" "vpc" {
  name = local.hname
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnet" {
  network       = google_compute_network.vpc.name
  name          = local.hname
  ip_cidr_range = var.subnet_cidr
  region        = var.region

  secondary_ip_range {
    range_name    = format("%s-secondary1", local.hname)
    ip_cidr_range = var.secondary_cidr_pods
  }

  secondary_ip_range {
    range_name    = format("%s-secondary2", local.hname)
    ip_cidr_range = var.secondary_cidr_services
  }

  private_ip_google_access = true
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
