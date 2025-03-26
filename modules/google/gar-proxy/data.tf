data "google_project" "this" {}

data "google_client_config" "this" {}

data "google_secret_manager_secret_version_access" "gar_proxy_username" {
  for_each = local.registries_with_credentials
  secret   = each.value.username_secret
}

data "google_secret_manager_secret_version_access" "gar_proxy_access_token" {
  for_each = local.registries_with_credentials
  secret   = each.value.access_token_secret
}