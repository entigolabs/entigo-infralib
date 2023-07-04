#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/.."
cd $SCRIPTPATH || exit 1

docker run -it --rm -v "$(pwd)":"/data" ghcr.io/terraform-linters/tflint-bundle:v0.46.1.1 
