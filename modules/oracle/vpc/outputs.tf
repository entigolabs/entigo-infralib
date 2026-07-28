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

# Single-value alias of public_subnets[0] - the agent's k8s Helm value templating
# (agent_input_oracle.yaml) has no way to index into a list output the way Terraform
# module inputs are auto-wired, so anything needing exactly one subnet OCID as a plain
# string (e.g. modules/k8s/oci-native-ingress-controller's IngressClassParameters) needs
# a dedicated singular output instead.
output "public_subnet_id" {
  value = element(oci_core_subnet.public[*].id, 0)
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
