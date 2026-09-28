#!/bin/bash
source ../../../common/generate_config.sh
docker run --rm -v "$(pwd)":"/data" $TFLINT_IMAGE --call-module-type=none --disable-rule terraform_required_providers
