#!/bin/bash
source ../../../common/generate_config.sh
# Newer than TFLINT_IMAGE: v0.50.3 can't parse provider functions (provider::azurerm::parse_resource_id)
docker run --rm -v "$(pwd)":"/data" ghcr.io/terraform-linters/tflint:v0.64.0 --call-module-type=none --disable-rule terraform_required_providers
