#!/bin/bash
# Mirrors this module's provider packages and pins them by digest, see ../provider-images.sh:
#   pullpush.sh digests   write values-images.yaml (after changing VERSION or PROVIDERS)
#   pullpush.sh           copy the packages to the mirrors

VERSION="v0.16.2"
REGISTRY="xpkg.upbound.io/crossplane-contrib"
PROVIDERS="provider-sql"

source "$(dirname "$0")/../provider-images.sh"
