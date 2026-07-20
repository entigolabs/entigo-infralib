output "vpc_id" {
  value = oci_core_vcn.this.id
}

output "vpc_name" {
  value = oci_core_vcn.this.display_name
}

output "vpc_cidr" {
  value = var.vpc_cidr
}

output "public_subnets" {
  value = oci_core_subnet.public[*].id
}

output "private_subnets" {
  value = oci_core_subnet.private[*].id
}

output "intra_subnets" {
  value = oci_core_subnet.intra[*].id
}

output "database_subnets" {
  value = oci_core_subnet.database[*].id
}

output "public_subnet_cidrs" {
  value = oci_core_subnet.public[*].cidr_block
}

output "private_subnet_cidrs" {
  value = oci_core_subnet.private[*].cidr_block
}

output "intra_subnet_cidrs" {
  value = oci_core_subnet.intra[*].cidr_block
}

output "database_subnet_cidrs" {
  value = oci_core_subnet.database[*].cidr_block
}

output "internet_gateway_id" {
  value = one(oci_core_internet_gateway.this[*].id)
}

output "nat_gateway_id" {
  value = one(oci_core_nat_gateway.this[*].id)
}

output "service_gateway_id" {
  value = one(oci_core_service_gateway.this[*].id)
}
