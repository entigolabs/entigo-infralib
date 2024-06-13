locals {
  #First range
  public_subnets = var.public_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 0)] : var.public_subnets
  intra_subnets  = var.intra_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 1)] : var.intra_subnets

  #second range
  private_subnets = var.private_subnets == null ? [cidrsubnet(var.vpc_cidr, 1, 0)] : var.private_subnets

  #third range
  database_subnets = var.database_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1),2,2)] : var.database_subnets

}

 
resource "google_compute_network" "vpc" {
  name = local.hname
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "private" {
  count = length(local.private_subnets)
  network       = google_compute_network.vpc.name
  name          = "${local.hname}-private-${count.index}"
  ip_cidr_range =  cidrsubnet(local.private_subnets[count.index], 2, 0)

  secondary_ip_range {
    range_name    = format("%s-pods", local.hname)
    ip_cidr_range = cidrsubnet(local.private_subnets[count.index], 1, 1)
  }

  secondary_ip_range {
    range_name    = format("%s-services", local.hname)
    ip_cidr_range = cidrsubnet(local.private_subnets[count.index], 2, 1)
  }

  private_ip_google_access = true
}

resource "google_compute_subnetwork" "public" {
  count = length(local.public_subnets)
  network       = google_compute_network.vpc.name
  name          = "${local.hname}-public-${count.index}"
  ip_cidr_range = local.public_subnets[count.index]
  purpose       = "REGIONAL_MANAGED_PROXY"
  role          = count.index == 0 ? "ACTIVE" : "BACKUP"

  private_ip_google_access = false
}

resource "google_compute_subnetwork" "intra" {
  count = length(local.intra_subnets)
  network       = google_compute_network.vpc.name
  name          = "${local.hname}-intra-${count.index}"
  ip_cidr_range    = local.intra_subnets[count.index]
}

resource "google_compute_subnetwork" "database" {
  count = length(local.database_subnets)
  network       = google_compute_network.vpc.name
  name          = "${local.hname}-database-${count.index}"
  ip_cidr_range    = local.database_subnets[count.index]
}

module "cloud_nat" {
  count                    = var.enable_nat_gateway ? 1 : 0
  source                             = "terraform-google-modules/cloud-nat/google"
  version                            = "5.2.0"
  project_id                         = data.google_client_config.this.project
  region                             = data.google_client_config.this.region
  router                             = google_compute_router.router.name
  name                               = local.hname
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

resource "google_compute_router" "router" {
  name    = local.hname
  network = google_compute_network.vpc.name
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
