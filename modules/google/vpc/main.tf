locals {
  # First range
  public_subnets = var.public_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 0)] : var.public_subnets
  intra_subnets  = var.intra_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 1)] : var.intra_subnets

  # Second range
  private_subnets = var.private_subnets == null ? [cidrsubnet(var.vpc_cidr, 1, 0)] : var.private_subnets

  # Third range
  database_subnets = var.database_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 2)] : var.database_subnets
}

resource "google_compute_network" "vpc" {
  name                    = var.prefix
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "private" {
  count   = length(local.private_subnets)
  network = google_compute_network.vpc.name
  name    = try(var.private_subnet_names[count.index], "${var.prefix}-private-${count.index}")

  ip_cidr_range = cidrsubnet(local.private_subnets[count.index], 2, 0)

  secondary_ip_range {
    range_name    = format("%s-pods", var.prefix)
    ip_cidr_range = cidrsubnet(local.private_subnets[count.index], 1, 1)
  }

  secondary_ip_range {
    range_name    = format("%s-services", var.prefix)
    ip_cidr_range = cidrsubnet(local.private_subnets[count.index], 2, 1)
  }

  private_ip_google_access = true
}

resource "google_compute_subnetwork" "public" {
  count         = length(local.public_subnets)
  network       = google_compute_network.vpc.name
  name          = try(var.public_subnet_names[count.index], "${var.prefix}-public-${count.index}")
  ip_cidr_range = local.public_subnets[count.index]
  purpose       = "REGIONAL_MANAGED_PROXY"
  role          = count.index == 0 ? "ACTIVE" : "BACKUP"

  private_ip_google_access = false
}

resource "google_compute_subnetwork" "intra" {
  count         = length(local.intra_subnets)
  network       = google_compute_network.vpc.name
  name          = try(var.intra_subnet_names[count.index], "${var.prefix}-intra-${count.index}")
  ip_cidr_range = local.intra_subnets[count.index]
}

resource "google_compute_subnetwork" "database" {
  count         = length(local.database_subnets)
  network       = google_compute_network.vpc.name
  name          = try(var.database_subnet_names[count.index], "${var.prefix}-database-${count.index}")
  ip_cidr_range = local.database_subnets[count.index]
}

resource "google_compute_address" "cloud_nat" {
  count = var.enable_nat_gateway ? var.nat_static_ip_count : 0
  name  = "${var.prefix}-nat-gateway-${count.index}"
}

# https://github.com/terraform-google-modules/terraform-google-cloud-nat
module "cloud_nat" {
  count                               = var.enable_nat_gateway ? 1 : 0
  source                              = "terraform-google-modules/cloud-nat/google"
  version                             = "7.0.0"
  project_id                          = data.google_client_config.this.project
  region                              = data.google_client_config.this.region
  router                              = google_compute_router.router.name
  name                                = var.prefix
  source_subnetwork_ip_ranges_to_nat  = "ALL_SUBNETWORKS_ALL_IP_RANGES"
  nat_ips                             = google_compute_address.cloud_nat[*].self_link
  enable_endpoint_independent_mapping = var.nat_enable_endpoint_independent_mapping
  enable_dynamic_port_allocation      = var.nat_enable_dynamic_port_allocation
  min_ports_per_vm                    = var.nat_min_ports_per_vm
  max_ports_per_vm                    = var.nat_max_ports_per_vm
  log_config_enable                   = var.nat_log_config_enable
  log_config_filter                   = var.nat_log_config_filter
  icmp_idle_timeout_sec               = var.nat_icmp_idle_timeout_sec
  tcp_established_idle_timeout_sec    = var.nat_tcp_established_idle_timeout_sec
  tcp_transitory_idle_timeout_sec     = var.nat_tcp_transitory_idle_timeout_sec
  tcp_time_wait_timeout_sec           = var.nat_tcp_time_wait_timeout_sec
  udp_idle_timeout_sec                = var.nat_udp_idle_timeout_sec
}

resource "google_compute_router" "router" {
  name    = var.prefix
  network = google_compute_network.vpc.name
}
