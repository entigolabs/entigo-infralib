#!/bin/bash
# Warm the Go module and build caches of the test image so that module tests
# download nothing at run time.
#
# run_tests() in images/tools/functions-common.sh sets up every module test the
# same way: init a module in /app, require entigo-infralib-common replaced by
# the bind-mounted /common, go mod tidy, go test. Which module versions tidy
# settles on, and which extra modules it fetches for go.sum checksums, depends
# on the exact set of packages the test imports. A bigger import set can pick a
# higher version of a test-only helper than a real module test does, and the
# lower one would then still be downloaded. So this does two things:
#
#   1. tidy + compile /app/test, which imports the union of everything the
#      module tests import, to fill the build cache once for all of them;
#   2. tidy + compile once per distinct import set found in the real module
#      tests under /warm (copied from modules/*/*/test at build time), so the
#      module cache holds exactly what each of them resolves.
set -euo pipefail
export GOMAXPROCS=2

setup_module() {
  go mod init github.com/entigolabs/entigo-infralib >/dev/null 2>&1
  go mod edit -require github.com/entigolabs/entigo-infralib-common@v0.0.0 \
    -replace github.com/entigolabs/entigo-infralib-common=/common
  go mod tidy
}

cd /common && go mod download -x

cd /app && setup_module && (cd test && go test -timeout 5m)

declare -A seen
count=0
for dir in $(find /warm -type d -name test | sort); do
  key=$(sed -nE 's/^\s*([A-Za-z0-9_.]+\s+)?"([^"]+)".*/\2/p' "$dir"/*_test.go | sort -u | md5sum | cut -d' ' -f1)
  if [ -n "${seen[$key]:-}" ]; then continue; fi
  seen[$key]=1
  count=$((count + 1))
  echo "Warming cache for import set of ${dir#/warm/}"
  (cd "$(dirname "$dir")" && setup_module && cd test && go test -c -o /dev/null .)
done
echo "Warmed ${count} distinct import sets from $(find /warm -type d -name test | wc -l) module tests"

find /go/pkg/mod -type f -name examples-1.json -delete
