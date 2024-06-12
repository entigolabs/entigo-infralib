output "vpc_id" {
  value = google_compute_network.vpc.id
}

output "subnet_cidr" {
  value = google_compute_subnetwork.subnet.ip_cidr_range
}

output "subnet_cidr_pods" {
  value = google_compute_subnetwork.subnet.secondary_ip_range[0].ip_cidr_range
}

output "subnet_cidr_services" {
  value = google_compute_subnetwork.subnet.secondary_ip_range[1].ip_cidr_range
}

output "nat_name" {
  value = module.cloud_nat.name
}