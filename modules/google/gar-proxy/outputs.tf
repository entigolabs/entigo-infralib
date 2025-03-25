output "hub_registry" {
  value = "${data.google_client_config.this.region}-docker.pkg.dev/${data.google_client_config.this.project}/${google_artifact_registry_repository.gar_proxy["hub"].repository_id}"
}

output "ghcr_registry" {
  value = "${data.google_client_config.this.region}-docker.pkg.dev/${data.google_client_config.this.project}/${google_artifact_registry_repository.gar_proxy["ghcr"].repository_id}"
}

output "gcr_registry" {
  value = "${data.google_client_config.this.region}-docker.pkg.dev/${data.google_client_config.this.project}/${google_artifact_registry_repository.gar_proxy["gcr"].repository_id}"
}

output "ecr_registry" {
  value = "${data.google_client_config.this.region}-docker.pkg.dev/${data.google_client_config.this.project}/${google_artifact_registry_repository.gar_proxy["ecr"].repository_id}"
}

output "quay_registry" {
  value = "${data.google_client_config.this.region}-docker.pkg.dev/${data.google_client_config.this.project}/${google_artifact_registry_repository.gar_proxy["quay"].repository_id}"
}

output "k8s_registry" {
  value = "${data.google_client_config.this.region}-docker.pkg.dev/${data.google_client_config.this.project}/${google_artifact_registry_repository.gar_proxy["k8s"].repository_id}"
}

