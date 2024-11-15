terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "6.11.1"
    }
    random = {
      source = "hashicorp/random"
      version = "3.6.2"
    }
    
  }
}
