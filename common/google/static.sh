#!/bin/bash

docker run --rm -v "$(pwd)":"/data" ghcr.io/terraform-linters/tflint:v0.50.3