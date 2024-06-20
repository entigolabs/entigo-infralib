resource "google_project_iam_member" "crossplane" {
  member  = local.member
  role    = "roles/editor"
  project = data.google_client_config.this.project
}