terraform {
  required_version = ">= 1.11"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "5.10.0"
    }
    azapi = {
      source  = "Azure/azapi"
      version = "2.13.0"
    }
  }
}
