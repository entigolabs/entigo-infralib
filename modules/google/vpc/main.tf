
resource "google_project_service" "compute" {
  service = "compute.googleapis.com"
}

 
resource "google_compute_network" "vpc" {
  name                                      = local.hname
  depends_on = [
    google_project_service.compute
  ]
}
