#!/usr/bin/env bash
# Vendors the community OpenTofu modules a module calls into the module
# itself, so that it carries them the way a chart carries charts/*.tgz and
# nothing is downloaded from the registry or GitHub at init time. Two files
# under <module-dir>/vendors, after Chart.yaml and Chart.lock:
#
#   vendors.yaml   written by hand (or by the update automation): what is
#                  wanted, a source and a version per module
#                    modules:
#                      - source: terraform-aws-modules/efs/aws
#                        version: 2.2.1
#                  source is written the way OpenTofu writes a module source.
#                  Without a prefix it is a registry address
#                  (<namespace>/<name>/<provider>), the only kind resolved
#                  today. A prefixed one (git::https://...) is reserved for
#                  modules outside the registry: only resolve() has to learn
#                  it, the fetch and the lock already work for any git host.
#   vendors.lock   written by this script: what that resolved to, the
#                  directory name, repository, commit and a checksum of the
#                  vendored tree
#
#   ./vendor.sh <module-dir>...
#       Like "helm dependency update": downloads every module of vendors.yaml
#       in the version the registry serves, prunes what a module call never
#       reads, writes it to <module-dir>/vendors/<repository-name> and writes
#       vendors.lock. The module calls it with
#       source = "./vendors/<repository-name>" (optionally //modules/<sub>)
#       and no version. A version bump is an edit of vendors.yaml and this.
#   ./vendor.sh --check [<module-dir>...]
#       Offline: vendors.lock is what vendors.yaml asks for and every vendored
#       tree still matches its checksum (nobody edited it by hand). Without
#       arguments, every module of the repository.
#
# Needs curl, git and sha256sum.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
REGISTRY="${VENDOR_REGISTRY:-https://registry.opentofu.org}"
die() { echo "ERROR: $*" >&2; exit 1; }

# tree_checksum DIR: one sha256 over the names and contents of every file.
tree_checksum() {
  ( cd "$1" && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum | sha256sum | cut -d' ' -f1 )
}

# wanted FILE: "source version" per entry of a vendors.yaml or vendors.lock.
wanted() {
  awk '{ sub(/[[:space:]]+#.*$/, "") }
       $1 == "-" { $1 = ""; $0 = $0 }
       $1 == "source:"  { source = $2; gsub(/["\x27]/, "", source) }
       $1 == "version:" { version = $2; gsub(/["\x27]/, "", version); print source, version }' "$1"
}

# locked FILE: "name checksum" per entry of a vendors.lock.
locked() {
  awk '$1 == "-" && $2 == "name:" { name = $3 } $1 == "checksum:" { sub(/^sha256:/, "", $2); print name, $2 }' "$1"
}

check() {
  local rc=0 dir vendors name want got d
  [ $# -gt 0 ] || set -- "$ROOT"/modules/*/*/
  for dir in "$@"; do
    vendors="${dir%/}/vendors"
    [ -d "$vendors" ] || continue
    [ -f "$vendors/vendors.yaml" ] || { echo "FAIL $vendors has no vendors.yaml" >&2; rc=1; continue; }
    [ -f "$vendors/vendors.lock" ] || { echo "FAIL $vendors has no vendors.lock: run ./vendor.sh ${dir%/}" >&2; rc=1; continue; }
    if [ "$(wanted "$vendors/vendors.yaml" | LC_ALL=C sort)" != "$(wanted "$vendors/vendors.lock" | LC_ALL=C sort)" ]; then
      echo "FAIL $vendors/vendors.lock is not what vendors.yaml asks for: run ./vendor.sh ${dir%/}" >&2; rc=1; continue
    fi
    while read -r name want; do
      got=$(tree_checksum "$vendors/$name" 2>/dev/null || true)
      if [ "$got" = "$want" ]; then echo "ok   $vendors/$name"
      else echo "FAIL $vendors/$name does not match vendors.lock (edited by hand?)" >&2; rc=1; fi
    done < <(locked "$vendors/vendors.lock")
    # A directory the lock does not know is as wrong as a changed one.
    for d in "$vendors"/*/; do
      [ -d "$d" ] || continue
      locked "$vendors/vendors.lock" | grep -q "^$(basename "$d") " || { echo "FAIL ${d%/} is not in vendors.lock" >&2; rc=1; }
    done
  done
  return $rc
}

# resolve SOURCE VERSION: sets REPO (git URL) and COMMIT. The one place that
# knows where a kind of source lives; everything after it is plain git.
resolve() {
  local source=$1 version=$2 answer get ref
  case "$source" in
    *::*) die "$source: only registry sources (<namespace>/<name>/<provider>) are supported so far, not ${source%%::*}:: ones" ;;
  esac
  [[ "$source" =~ ^[A-Za-z0-9_-]+/[A-Za-z0-9_-]+/[A-Za-z0-9_-]+$ ]] || die "$source is not a registry source (<namespace>/<name>/<provider>, without //submodule)"

  # The registry answers a download request with where the version lives, in
  # the body ({"location": "git::https://<host>/<org>/<repo>?ref=<commit>"})
  # or, on registries that follow the older protocol, in X-Terraform-Get.
  answer=$(curl -sSfL -D - "$REGISTRY/v1/modules/$source/$version/download" | tr -d '\r') || die "$source $version is not in $REGISTRY"
  get=$(grep -oE 'git::https://[^"[:space:]]+' <<<"$answer" | tail -n1 || true)
  [[ "$get" =~ ^git::(https://[^?]+)\?ref=([^\&]+)$ ]] || die "unsupported download location for $source $version: '${get}'"
  REPO="${BASH_REMATCH[1]%.git}"; ref="${BASH_REMATCH[2]}"

  # The registry pins a commit; a tag is resolved to one, so that the record
  # survives a moved tag.
  if [[ "$ref" =~ ^[0-9a-f]{40}$ ]]; then COMMIT="$ref"
  else COMMIT=$(git ls-remote "$REPO" "refs/tags/$ref^{}" "refs/tags/$ref" | sort -k2 -r | head -n1 | cut -f1); fi
  [[ "$COMMIT" =~ ^[0-9a-f]{40}$ ]] || die "cannot resolve $ref in $REPO"
}

# fetch SOURCE VERSION STAGE: one module into STAGE/<repository-name>, its
# lock entry appended to STAGE/vendors.lock.
fetch() {
  local source=$1 version=$2 stage=$3 REPO COMMIT repo commit name src nested
  resolve "$source" "$version"
  repo=$REPO; commit=$COMMIT; name="$(basename "$repo")"

  src="$stage/$name"
  [ ! -e "$src" ] || die "vendors.yaml lists $name twice"
  # That one commit, over git: no host has to offer an archive download.
  git init -q "$src"
  git -C "$src" -c advice.detachedHead=false fetch -q --depth 1 "$repo" "$commit" || die "cannot fetch $commit from $repo"
  git -C "$src" -c advice.detachedHead=false checkout -q FETCH_HEAD

  # Pruned by exclusion, not by a list of wanted extensions: a module may read
  # any file it ships (templates, scripts, JSON), but never these.
  ( cd "$src"
    rm -rf .git .github examples test tests wrappers docs build autogen helpers
    rm -f CODEOWNERS Makefile Gemfile Gemfile.lock kitchen.yml metadata.yaml metadata.display.yaml
    find . -mindepth 2 -type f \( -name 'metadata.yaml' -o -name 'metadata.display.yaml' \) -delete
    find . -maxdepth 1 -type f -name '.*' -delete
    find . -type f -name '*.md' -delete
    find . -type l -delete
    find . -type d -empty -delete )
  [ -n "$(find "$src" -name '*.tf' -print -quit)" ] || die "$repo@$commit has no module files after pruning"
  ls "$src" | grep -qiE '^licen[cs]e' || die "$repo@$commit ships no licence file; it cannot be redistributed as is"

  # A nested registry or git call would still be downloaded at init time.
  nested=$(grep -rnE '^\s*source\s*=\s*"[^."]' --include='*.tf' "$src" | grep -vE 'source\s*=\s*"[A-Za-z0-9_-]+/[A-Za-z0-9_-]+"' || true)
  [ -z "$nested" ] || { echo "WARNING: $name calls modules that are not vendored; they need vendoring and a source rewrite:" >&2; echo "$nested" | sed "s#$src/#  #" >&2; }

  {
    echo "  - name: $name"
    echo "    source: $source"
    echo "    version: $version"
    echo "    repository: $repo"
    echo "    commit: $commit"
    echo "    checksum: sha256:$(tree_checksum "$src")"
  } >> "$stage/vendors.lock"
  echo "Vendored $source $version (${commit:0:12}) as $name: $(find "$src" -type f | wc -l) files, $(du -sk "$src" | cut -f1)K"
}

update() {
  local dir=${1%/} vendors stage source version d
  vendors="$dir/vendors"
  [ -f "$dir/main.tf" ] || die "$dir is not a module directory"
  [ -f "$vendors/vendors.yaml" ] || die "$vendors/vendors.yaml does not exist: list the modules to vendor there"
  [ -n "$(wanted "$vendors/vendors.yaml")" ] || die "$vendors/vendors.yaml lists no module (source and version)"

  # Everything is fetched next to the target first: a failed download leaves
  # the vendored trees and the lock as they were.
  stage=$(mktemp -d "$vendors/.stage.XXXXXX"); STAGE=$stage
  {
    echo "# Written by ./vendor.sh from vendors.yaml: what each module resolved to."
    echo "# Never edit this file or the directories next to it by hand."
    echo "modules:"
  } > "$stage/vendors.lock"
  while read -r source version; do fetch "$source" "$version" "$stage"; done < <(wanted "$vendors/vendors.yaml")

  for d in "$vendors"/*/; do [ ! -d "$d" ] || rm -rf "$d"; done
  mv "$stage"/*/ "$vendors/"
  mv "$stage/vendors.lock" "$vendors/vendors.lock"
  rm -rf "$stage"
}

# A failed update exits from inside a function; its stage must not stay behind.
STAGE=""; trap '[ -z "$STAGE" ] || rm -rf "$STAGE"' EXIT
if [ "${1:-}" = "--check" ]; then shift; check "$@"; exit $?; fi
[ $# -gt 0 ] || { sed -n '2,28p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//' >&2; exit 2; }
for dir in "$@"; do update "$dir"; done
