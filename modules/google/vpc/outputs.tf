output "vpc_id" {
  value = google_compute_network.vpc.id
}

output "private_subnets" {
  value = google_compute_subnetwork.private[*].id
}

output "public_subnets" {
  value = google_compute_subnetwork.public[*].id
}

output "intra_subnets" {
  value = google_compute_subnetwork.intra[*].id
}

output "database_subnets" {
  value = google_compute_subnetwork.database[*].id
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
