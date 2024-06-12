 
 resource "google_compute_network" "vpc" {
  name = "${local.hname}-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnet" {
  network       = google_compute_network.vpc.name
  name          = "${local.hname}-subnet"
  ip_cidr_range = var.subnet_cidr
  region        = var.region

  secondary_ip_range {
    range_name    = format("%s-pods", local.hname)
    ip_cidr_range = var.secondary_cidr_pods
  }

  secondary_ip_range {
    range_name    = format("%s-services", local.hname)
    ip_cidr_range = var.secondary_cidr_services
  }

  private_ip_google_access = true
}

module "cloud_nat" {
  source                             = "terraform-google-modules/cloud-nat/google"
  version                            = "~> 5.0"
  project_id                         = var.project_id
  region                             = var.region
  router                             = google_compute_router.router.name
  name                               = "${local.hname}-nat-config"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

resource "google_compute_router" "router" {
  project = var.project_id
  name    = "${local.hname}-nat-router"
  network = google_compute_network.vpc.name
  region  = var.region
}

# Secrets
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