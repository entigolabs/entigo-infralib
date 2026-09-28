#!/bin/bash
source ../../../common/generate_config.sh
docker run --rm -v "$(dirname "$(pwd)")":"/data" $TFLINT_IMAGE --chdir="$(basename "$(pwd)")" --disable-rule terraform_required_providers
