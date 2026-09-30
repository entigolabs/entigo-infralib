terraform {
  required_version = ">= 1.11"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "5.7.0"
    }
    azapi = {
      source  = "Azure/azapi"
      version = "2.12.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.14.2"
    }
  }
}
