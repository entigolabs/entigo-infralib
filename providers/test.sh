#!/usr/bin/env bash
# Static check of the provider templates: every module that pins a provider
# pins the same version, and the templates lint with tflint using those
# versions. Needs docker only.
set -euo pipefail
cd "$(dirname "$0")"

TFLINT_IMAGE="${TFLINT_IMAGE:-ghcr.io/terraform-linters/tflint:v0.64.0}"

cp base.tf test_base.tf
trap 'rm -rf test_base.tf tmp_tf' EXIT
for file in *.tf; do
  case "$file" in base.tf|test_base.tf) continue ;; esac
  providername=${file%.tf}
  echo "Version unity check of $file"
  lastversion=""
  for verfile in $(find ../modules/ -name versions.tf); do
    while read -r versionfound; do
      [ -n "$versionfound" ] || continue
      echo "Found $providername version $versionfound in $verfile"
      if ! [[ "$versionfound" =~ ^[0-9]+(\.[0-9]+)*$ ]]; then
        echo "Skipping $verfile, $providername version $versionfound is a constraint and not a pinned version"
        continue
      fi
      if [ -z "$lastversion" ]; then
        lastversion=$versionfound
      elif [ "$lastversion" != "$versionfound" ]; then
        echo "Version mismatch for $providername $lastversion != $versionfound in $verfile"
        exit 1
      fi
    done < <(awk -v keyword="$providername" '$1 == keyword && $2 == "=" { while(getline > 0 && $1 != "}") { if($1 == "version") { sub(/^[[:space:]]*version[[:space:]]*=[[:space:]]*/, ""); sub(/[[:space:]]*$/, ""); gsub(/"/, ""); print; } } }' "$verfile")
  done
  case "$providername" in
    helmaws|helmgoogle) modulename=helm ;;
    *) modulename=$providername ;;
  esac
  case "$providername" in
    oci) sourceorg=oracle ;;
    azapi) sourceorg=Azure ;;
    *) sourceorg=hashicorp ;;
  esac
  awk -v providername="$providername" -v modulename="$modulename" -v sourceorg="$sourceorg" -v lastversion="$lastversion" \
    '/required_providers {/ { print; print "    " providername " = {\n      source  = \"" sourceorg "/" modulename "\"\n      version = \"" lastversion "\"\n    }"; next }1' test_base.tf > tmp && mv tmp test_base.tf
done

rm -rf tmp_tf && mkdir tmp_tf
cp ./*.tf tmp_tf/
rm tmp_tf/base.tf
docker run --rm --user "$(id -u):$(id -g)" -e HOME=/tmp -v "$(pwd)/tmp_tf":/data "$TFLINT_IMAGE"
echo "Providers OK"
