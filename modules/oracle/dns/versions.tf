terraform {
  required_version = ">= 1.5"
  required_providers {
    oci = {
      source  = "oracle/oci"
      version = "8.23.0"
    }
    # Let's Encrypt client for the zone-wide wildcard certificate (certificate.tf). The
    # provider embeds lego, which is where the OCI DNS-01 solver comes from - see
    # https://go-acme.github.io/lego/dns/oraclecloud/.
    acme = {
      source  = "vancluever/acme"
      version = "2.48.3"
    }
    tls = {
      source  = "hashicorp/tls"
      version = ">= 4.0.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.0"
    }
  }
}
