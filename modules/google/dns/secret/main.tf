variable "prefix" {
  type = string
}

variable "key" {
  type = string
}

variable "value" {
}


resource "google_secret_manager_secret" "secret" {
  secret_id = "entigo-infralib-${var.prefix}-${var.key}"

  annotations = {
    product = "entigo-infralib"
    prefix = var.prefix
    parameter = "secret"
  }
  
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "secret" {
  secret = google_secret_manager_secret.secret.id
  secret_data = try("\"${join("\",\"", var.value)}\"", var.value)
}  
