terraform {
  required_version = ">= 1.11"
  required_providers {
    time = {
      source  = "hashicorp/time"
      version = "0.14.2"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "5.9.0"
    }
    azapi = {
      source  = "Azure/azapi"
      version = "2.13.0"
    }
  }
}
