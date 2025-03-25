locals {
  registries = {
    hub  = { uri = "registry-1.docker.io", username = var.hub_username, access_token_secret_name = var.hub_access_token_secret_name }
    ghcr = { uri = "ghcr.io", username = var.ghcr_username, access_token_secret_name = var.ghcr_access_token_secret_name }
    gcr  = { uri = "gcr.io", username = var.gcr_username, access_token_secret_name = var.gcr_access_token_secret_name }
    ecr  = { uri = "public.ecr.aws", username = "", access_token_secret_name = "" }
    quay = { uri = "quay.io", username = "", access_token_secret_name = "" }
    k8s  = { uri = "registry.k8s.io", username = "", access_token_secret_name = "" }
  }

  registries_with_credentials = { for k, v in local.registries : k => v if v.username != "" && v.access_token_secret_name != "" }
}

resource "google_secret_manager_secret_iam_member" "gar_proxy" {
  for_each  = local.registries_with_credentials
  secret_id = each.value.access_token_secret_name
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:service-${data.google_project.this.number}@gcp-sa-artifactregistry.iam.gserviceaccount.com"
}

resource "google_artifact_registry_repository" "gar_proxy" {
  for_each      = local.registries
  depends_on    = [google_secret_manager_secret_iam_member.gar_proxy]
  repository_id = "${substr(var.prefix, 0, 50)}-${each.key}"
  format        = "DOCKER"
  mode          = "REMOTE_REPOSITORY"

  remote_repository_config {
    common_repository {
      uri = "https://${each.value.uri}"
    }

    dynamic "upstream_credentials" {
      for_each = each.value.username != "" && each.value.access_token_secret_name != "" ? [1] : []
      content {
        username_password_credentials {
          username                = each.value.username
          password_secret_version = data.google_secret_manager_secret_version_access.gar_proxy[each.key].name
        }
      }
    }
  }

  vulnerability_scanning_config {
    enablement_config = "DISABLED"
  }

  cleanup_policies {
    id     = "delete-untagged-older-than-7d"
    action = "DELETE"
    condition {
      tag_state  = "UNTAGGED"
      older_than = "7d"
    }
  }

  cleanup_policies {
    id     = "delete-all-older-than-90d"
    action = "DELETE"
    condition {
      older_than = "90d"
    }
  }

}
