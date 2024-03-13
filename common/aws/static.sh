#!/bin/bash

docker run --rm -v "$(pwd)":"/data" ghcr.io/terraform-linters/tflint-bundle:v0.46.1.1
