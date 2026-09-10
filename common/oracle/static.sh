#!/bin/bash
source ../../../common/generate_config.sh
docker run --rm -v "$(pwd)":"/data" $TFLINT_IMAGE --disable-rule terraform_required_providers
