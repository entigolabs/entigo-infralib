locals {
  registries = {
    hub  = { uri = "registry-1.docker.io", username_secret = var.hub_username_secret, access_token_secret = var.hub_access_token_secret }
    ghcr = { uri = "ghcr.io", username_secret = var.ghcr_username_secret, access_token_secret = var.ghcr_access_token_secret }
    gcr  = { uri = "gcr.io", username_secret = var.gcr_username_secret, access_token_secret = var.gcr_access_token_secret }
    ecr  = { uri = "public.ecr.aws", username_secret = var.ecr_username_secret, access_token_secret = var.ecr_access_token_secret }
    quay = { uri = "quay.io", username_secret = var.quay_username_secret, access_token_secret = var.quay_access_token_secret }
    k8s  = { uri = "registry.k8s.io", username_secret = "", access_token_secret = "" }
  }

  registries_with_credentials = { for k, v in local.registries : k => v if v.username_secret != "" && v.access_token_secret != "" }
}

resource "google_secret_manager_secret_iam_member" "gar_proxy" {
  for_each  = local.registries_with_credentials
  secret_id = each.value.access_token_secret
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
      for_each = each.value.username_secret != "" && each.value.access_token_secret != "" ? [1] : []
      content {
        username_password_credentials {
          username                = data.google_secret_manager_secret_version_access.gar_proxy_username[each.key].secret_data
          password_secret_version = data.google_secret_manager_secret_version_access.gar_proxy_access_token[each.key].name
        }
      }
    }
  }

  vulnerability_scanning_config {
    enablement_config = var.enablement_config
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
