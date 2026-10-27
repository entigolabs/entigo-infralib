#!/usr/bin/env bash
# Entry point of this repository's module tests. Copy of
# entigo-infralib-test/templates/test.sh: fetches the pinned version of the
# orchestrator from the entigolabs/entigo-infralib-test-cli image into the
# git-ignored .infralib-test/ and runs it. The only line to edit is the version.
set -euo pipefail

INFRALIB_TEST_VERSION="${INFRALIB_TEST_VERSION:-v0.8.8}"

IMAGE="${INFRALIB_TEST_IMAGE_PREFIX:-entigolabs/entigo-infralib-test-}cli:${INFRALIB_TEST_VERSION}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
CACHE="$ROOT/.infralib-test/scripts-${INFRALIB_TEST_VERSION}"
case "$INFRALIB_TEST_VERSION" in latest|dev) rm -rf "$CACHE" ;; esac
if [ ! -x "$CACHE/infralib-test.sh" ]; then
  docker pull -q "$IMAGE" >/dev/null 2>&1 || docker image inspect "$IMAGE" >/dev/null 2>&1 || { echo "cannot pull $IMAGE" >&2; exit 1; }
  id=$(docker create "$IMAGE")
  mkdir -p "$ROOT/.infralib-test" && rm -rf "$CACHE.tmp"
  docker cp "$id:/opt/infralib-test/scripts" "$CACHE.tmp" && docker rm "$id" >/dev/null && mv "$CACHE.tmp" "$CACHE"
fi
INFRALIB_TEST_VERSION="$INFRALIB_TEST_VERSION" exec "$CACHE/infralib-test.sh" "$@"
