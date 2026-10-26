#!/usr/bin/env bash
# Test this module: hands over to the repository's test.sh with this module selected (./test.sh --help).
MODULE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
exec "$MODULE/../../../test.sh" "$@" "modules/$(basename "$(dirname "$MODULE")")/$(basename "$MODULE")"
