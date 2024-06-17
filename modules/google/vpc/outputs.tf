output "vpc_id" {
  value = google_compute_network.vpc.id
}

output "vpc_name" {
  value = google_compute_network.vpc.name
}

output "private_subnets" {
  value = google_compute_subnetwork.private[*].name
}

output "private_subnets_pods" {
  value = google_compute_subnetwork.private[*].secondary_ip_range[0].range_name
}

output "private_subnets_services" {
  value = google_compute_subnetwork.private[*].secondary_ip_range[1].range_name
}

output "public_subnets" {
  value = google_compute_subnetwork.public[*].name
}

output "intra_subnets" {
  value = google_compute_subnetwork.intra[*].name
}

output "database_subnets" {
  value = google_compute_subnetwork.database[*].name
}

output "private_subnet_cidrs" {
  value = google_compute_subnetwork.private[*].ip_cidr_range
}

output "private_subnet_cidrs_pods" {
  value = google_compute_subnetwork.private[*].secondary_ip_range[0].ip_cidr_range
}

output "private_subnet_cidrs_services" {
  value = google_compute_subnetwork.private[*].secondary_ip_range[1].ip_cidr_range
}

output "public_subnet_cidrs" {
  value = google_compute_subnetwork.public[*].ip_cidr_range
}

output "database_subnet_cidrs" {
  value = google_compute_subnetwork.database[*].ip_cidr_range
}

output "intra_subnet_cidrs" {
  value = google_compute_subnetwork.intra[*].ip_cidr_range
}

output "router_id" {
  value = google_compute_router.router.id
}

output "nat_name" {
  value = module.cloud_nat[0].name
}
