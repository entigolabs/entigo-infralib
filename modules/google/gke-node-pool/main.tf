# resource "google_container_node_pool" "this" {
#   name = substr(var.prefix, 0, 40)
#   cluster = var.cluster_name

#   node_locations =  length(var.node_locations) > 0 ? var.node_locations : data.google_compute_zones.this.names
  
#   node_config {
#     machine_type    = var.instance_type
#     disk_size_gb    = var.volume_size
#     disk_type       = var.volume_type
#     image_type      = "COS_CONTAINERD"
#     service_account = var.service_account
#     oauth_scopes = [
#       "https://www.googleapis.com/auth/cloud-platform"
#     ]
#     spot = var.spot_nodes
#   }

#   autoscaling {
#     location_policy = var.location_policy
#     min_node_count = var.min_size
#     max_node_count = var.max_size
#   }

#   max_pods_per_node  = var.max_pods
#   initial_node_count = var.min_size

#   management {
#     auto_repair  = true
#     auto_upgrade = false
#   }

#   upgrade_settings {
#     max_surge       = 1
#     max_unavailable = 0
#   }

# }


module "gke_node_pool" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/gke-node-pool"
  version = "41.0.2"

  name = substr(var.prefix, 0, 40)

  cluster = var.cluster_name

  initial_node_count = var.min_size

  kubernetes_version = data.google_container_engine_versions.this.release_channel_latest_version["STABLE"]

  location = var.cluster_region

  node_locations = length(var.node_locations) > 0 ? var.node_locations : data.google_compute_zones.this.names

  max_pods_per_node = var.max_pods
  
  management = {
    auto_repair  = true
    auto_upgrade = false
  }

  autoscaling = {
    min_node_count = var.min_size
    max_node_count = var.max_size
    total_min_node_count = var.min_size
    total_max_node_count = var.max_size
    location_policy = var.location_policy
  }

  network_config = {
    
  }

  node_config = {
    
  }

  placement_policy = {

  }

  queued_provisioning	= {

  }

  timeouts = {
   create = "45m"
   update = "45m"
   delete = "45m"
  }

  upgrade_settings = {
    
  }
 
}
