locals {
  # Same default subnet split as the Google VPC module.
  # First range
  public_subnets = var.public_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 0)] : var.public_subnets
  intra_subnets  = var.intra_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 1)] : var.intra_subnets

  # Second range
  private_subnets = var.private_subnets == null ? [cidrsubnet(var.vpc_cidr, 1, 0)] : var.private_subnets

  # Third range
  database_subnets = var.database_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 2)] : var.database_subnets

  # Fourth range - pods, for OKE's VCN-native pod networking (see modules/oracle/oke).
  # Every pod gets a real VCN IP off this subnet, the way EKS pods get VPC IPs. Takes the
  # previously unused fourth quarter of the second range rather than resizing any existing
  # tier, so adding this cannot renumber a deployed VCN. A /19 out of the default /16 holds
  # ~8190 addresses; OKE reserves 31 per worker node (one secondary VNIC's worth), so that
  # is room for roughly 260 nodes.
  pod_subnets = var.pod_subnets == null ? [cidrsubnet(cidrsubnet(var.vpc_cidr, 1, 1), 2, 3)] : var.pod_subnets

  services_cidr = data.oci_core_services.all.services[0].cidr_block

  # Route rules per subnet tier. Conditionals short-circuit, so the [0] index is
  # only evaluated when the matching gateway is created.
  #
  # The public table intentionally omits the service gateway: OCI rejects an
  # internet gateway and an "All Services" service gateway target in the same
  # route table. Public subnets reach OCI services via the internet gateway.
  public_routes = var.enable_internet_gateway ? [{
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_internet_gateway.this[0].id
  }] : []

  private_routes = concat(
    var.enable_nat_gateway ? [{
      destination       = "0.0.0.0/0"
      destination_type  = "CIDR_BLOCK"
      network_entity_id = oci_core_nat_gateway.this[0].id
    }] : [],
    var.enable_service_gateway ? [{
      destination       = local.services_cidr
      destination_type  = "SERVICE_CIDR_BLOCK"
      network_entity_id = oci_core_service_gateway.this[0].id
    }] : [],
  )

  # Intra and database tiers stay isolated from the internet: service gateway only.
  internal_routes = var.enable_service_gateway ? [{
    destination       = local.services_cidr
    destination_type  = "SERVICE_CIDR_BLOCK"
    network_entity_id = oci_core_service_gateway.this[0].id
  }] : []
}

resource "oci_core_vcn" "this" {
  compartment_id = var.compartment_id
  cidr_blocks    = [var.vpc_cidr]
  display_name   = var.prefix
  dns_label      = var.dns_label
}

resource "oci_core_internet_gateway" "this" {
  count          = var.enable_internet_gateway ? 1 : 0
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${var.prefix}-igw"
}

resource "oci_core_nat_gateway" "this" {
  count          = var.enable_nat_gateway ? 1 : 0
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${var.prefix}-nat"
}

resource "oci_core_service_gateway" "this" {
  count          = var.enable_service_gateway ? 1 : 0
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${var.prefix}-sgw"

  services {
    service_id = data.oci_core_services.all.services[0].id
  }
}

resource "oci_core_route_table" "public" {
  count          = length(local.public_subnets) > 0 ? 1 : 0
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${var.prefix}-public"

  dynamic "route_rules" {
    for_each = local.public_routes
    content {
      destination       = route_rules.value.destination
      destination_type  = route_rules.value.destination_type
      network_entity_id = route_rules.value.network_entity_id
    }
  }
}

resource "oci_core_route_table" "private" {
  # Shared with the pod subnet, which needs the same NAT + service gateway egress.
  count          = length(local.private_subnets) + length(local.pod_subnets) > 0 ? 1 : 0
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${var.prefix}-private"

  dynamic "route_rules" {
    for_each = local.private_routes
    content {
      destination       = route_rules.value.destination
      destination_type  = route_rules.value.destination_type
      network_entity_id = route_rules.value.network_entity_id
    }
  }
}

resource "oci_core_route_table" "internal" {
  count          = length(local.intra_subnets) + length(local.database_subnets) > 0 ? 1 : 0
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.this.id
  display_name   = "${var.prefix}-internal"

  dynamic "route_rules" {
    for_each = local.internal_routes
    content {
      destination       = route_rules.value.destination
      destination_type  = route_rules.value.destination_type
      network_entity_id = route_rules.value.network_entity_id
    }
  }
}

resource "oci_core_subnet" "public" {
  count                      = length(local.public_subnets)
  compartment_id             = var.compartment_id
  vcn_id                     = oci_core_vcn.this.id
  cidr_block                 = local.public_subnets[count.index]
  display_name               = try(var.public_subnet_names[count.index], "${var.prefix}-public-${count.index}")
  route_table_id             = oci_core_route_table.public[0].id
  prohibit_public_ip_on_vnic = false
}

resource "oci_core_subnet" "private" {
  count                      = length(local.private_subnets)
  compartment_id             = var.compartment_id
  vcn_id                     = oci_core_vcn.this.id
  cidr_block                 = local.private_subnets[count.index]
  display_name               = try(var.private_subnet_names[count.index], "${var.prefix}-private-${count.index}")
  route_table_id             = oci_core_route_table.private[0].id
  prohibit_public_ip_on_vnic = true
}

# Pods under OKE's VCN-native CNI. Shares the private tier's route table so pods reach OCI
# services through the service gateway and the internet through NAT. No availability_domain
# is set, which is what makes it regional - OKE requires that of a pod subnet.
resource "oci_core_subnet" "pod" {
  count                      = length(local.pod_subnets)
  compartment_id             = var.compartment_id
  vcn_id                     = oci_core_vcn.this.id
  cidr_block                 = local.pod_subnets[count.index]
  display_name               = try(var.pod_subnet_names[count.index], "${var.prefix}-pod-${count.index}")
  route_table_id             = oci_core_route_table.private[0].id
  prohibit_public_ip_on_vnic = true
}

resource "oci_core_subnet" "intra" {
  count                      = length(local.intra_subnets)
  compartment_id             = var.compartment_id
  vcn_id                     = oci_core_vcn.this.id
  cidr_block                 = local.intra_subnets[count.index]
  display_name               = try(var.intra_subnet_names[count.index], "${var.prefix}-intra-${count.index}")
  route_table_id             = oci_core_route_table.internal[0].id
  prohibit_public_ip_on_vnic = true
}

resource "oci_core_subnet" "database" {
  count                      = length(local.database_subnets)
  compartment_id             = var.compartment_id
  vcn_id                     = oci_core_vcn.this.id
  cidr_block                 = local.database_subnets[count.index]
  display_name               = try(var.database_subnet_names[count.index], "${var.prefix}-database-${count.index}")
  route_table_id             = oci_core_route_table.internal[0].id
  prohibit_public_ip_on_vnic = true
}
