provider "oci" {
  # Auth and region come from the environment so the same config works both
  # in-container (auth=ResourcePrincipal via OCI_AUTH + OCI_RESOURCE_PRINCIPAL_*)
  # and locally (default ApiKey auth from ~/.oci/config or OCI_/TF_VAR_ env vars).
}
