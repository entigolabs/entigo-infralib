output "cluster_id" {
  value = oci_containerengine_cluster.this.id
}

output "cluster_name" {
  value = oci_containerengine_cluster.this.name
}

output "kubernetes_version" {
  value = oci_containerengine_cluster.this.kubernetes_version
}

output "public_endpoint" {
  value = oci_containerengine_cluster.this.endpoints[0].public_endpoint
}

output "private_endpoint" {
  value = oci_containerengine_cluster.this.endpoints[0].private_endpoint
}

output "kubernetes_endpoint" {
  value = oci_containerengine_cluster.this.endpoints[0].kubernetes
}

output "main_node_pool_id" {
  value = try(module.main[0].node_pool_id, "")
}

output "mon_node_pool_id" {
  value = try(module.mon[0].node_pool_id, "")
}

output "tools_node_pool_id" {
  value = try(module.tools[0].node_pool_id, "")
}

output "lb_nsg_id" {
  value = oci_core_network_security_group.lb.id
}

output "controllers_dynamic_group_name" {
  # Referenced by the per-app Crossplane Policy CRs in k8s modules' templates/oracle/
  # ("Allow dynamic-group <this> to ..." statements).
  value = oci_identity_dynamic_group.controllers.name
}

output "tenancy_id" {
  # Used by crossplane-oracle's Instance Principal credentials secret.
  value = data.oci_identity_compartment.this.compartment_id
}

output "region" {
  # The OCI provider has no "current region" data source; OCIDs of regional resources
  # embed the region identifier as the 4th dot-separated field
  # (ocid1.cluster.oc1.<region>.<hash>), so derive it from the cluster's own OCID.
  value = split(".", oci_containerengine_cluster.this.id)[3]
}

output "main_min_size" {
  # tostring: the agent renders numeric terraform outputs as %f floats ("1.000000"),
  # which breaks cluster-autoscaler's --nodes=<min>:<max>:<ocid> parsing (it then
  # falls back to the instance-pool implementation and crash-loops). Strings pass
  # through Go templating verbatim.
  value = tostring(var.oke_main_min_size)
}

output "main_max_size" {
  # tostring: the agent renders numeric terraform outputs as %f floats ("1.000000"),
  # which breaks cluster-autoscaler's --nodes=<min>:<max>:<ocid> parsing (it then
  # falls back to the instance-pool implementation and crash-loops). Strings pass
  # through Go templating verbatim.
  value = tostring(var.oke_main_max_size)
}

output "mon_min_size" {
  # tostring: the agent renders numeric terraform outputs as %f floats ("1.000000"),
  # which breaks cluster-autoscaler's --nodes=<min>:<max>:<ocid> parsing (it then
  # falls back to the instance-pool implementation and crash-loops). Strings pass
  # through Go templating verbatim.
  value = tostring(var.oke_mon_min_size)
}

output "mon_max_size" {
  # tostring: the agent renders numeric terraform outputs as %f floats ("1.000000"),
  # which breaks cluster-autoscaler's --nodes=<min>:<max>:<ocid> parsing (it then
  # falls back to the instance-pool implementation and crash-loops). Strings pass
  # through Go templating verbatim.
  value = tostring(var.oke_mon_max_size)
}

output "tools_min_size" {
  # tostring: the agent renders numeric terraform outputs as %f floats ("1.000000"),
  # which breaks cluster-autoscaler's --nodes=<min>:<max>:<ocid> parsing (it then
  # falls back to the instance-pool implementation and crash-loops). Strings pass
  # through Go templating verbatim.
  value = tostring(var.oke_tools_min_size)
}

output "tools_max_size" {
  # tostring: the agent renders numeric terraform outputs as %f floats ("1.000000"),
  # which breaks cluster-autoscaler's --nodes=<min>:<max>:<ocid> parsing (it then
  # falls back to the instance-pool implementation and crash-loops). Strings pass
  # through Go templating verbatim.
  value = tostring(var.oke_tools_max_size)
}
