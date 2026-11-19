#!/usr/bin/env python3
"""Garbage collection for ECR Public repositories, which have no lifecycle policies.

A manifest is kept when it is reachable from a tag:

  * it has a tag (a version, latest, dev, ...),
  * a kept image index lists it: the per-architecture images of a multi-arch
    tag and the SBOM and provenance attestations buildx adds to it,
  * its "subject" is a kept manifest: an SBOM or signature attached to it
    (oras attach, the OCI referrers of the manifest),
  * its only tags are cosign's sha256-<digest>.sig/.att/.sbom (or the
    referrers fallback tag sha256-<digest>) of a kept manifest.

Everything pushed within the grace period is kept as well, with whatever it
references: a build pushes every architecture untagged and tags the index
in a later job, so an image in the making looks like garbage. A manifest
that could not be read from the registry is kept too.

The rest is deleted, or only reported without --delete: untagged images
nothing refers to any more, and the signatures and attachments of images
that are gone.

    ecr-public-gc.py [--delete] [--grace-days N] [--largest N] [--alias ALIAS] REPOSITORY...

--largest N also reports the N largest repositories of the whole registry,
by the sum of their manifest sizes, after anything was deleted.

Needs the aws CLI with credentials for the registry's account
(ecr-public:DescribeImages, GetAuthorizationToken, with --largest
DescribeRepositories and with --delete BatchDeleteImage). Exit status 1 when a repository could not be processed or
a deletion failed.
"""
import argparse
import base64
import json
import os
import re
import subprocess
import sys
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timedelta, timezone

REGION = "us-east-1"  # the only region of the ECR Public API
REGISTRY = "https://public.ecr.aws"
INDEX_TYPES = (
    "application/vnd.oci.image.index.v1+json",
    "application/vnd.docker.distribution.manifest.list.v2+json",
)
ACCEPT = ", ".join(INDEX_TYPES + (
    "application/vnd.oci.image.manifest.v1+json",
    "application/vnd.docker.distribution.manifest.v2+json",
    "application/vnd.oci.artifact.manifest.v1+json",
))
# cosign's tags for the signature, attestation and SBOM of a digest, and the
# OCI referrers fallback tag, which has no suffix.
DIGEST_TAG = re.compile(r"^sha256-([0-9a-f]{64})(\.(sig|att|sbom))?$")


def aws(*args):
    out = subprocess.run(("aws",) + args + ("--region", REGION, "--output", "json"),
                         check=True, capture_output=True, text=True).stdout
    return json.loads(out) if out.strip() else {}


def list_images(repository):
    """Every manifest of the repository: digest -> tags, push time, size, media type."""
    images, token = {}, None
    while True:
        args = ["ecr-public", "describe-images", "--repository-name", repository, "--max-results", "1000"]
        if token:
            args += ["--next-token", token]
        page = aws(*args)
        for d in page.get("imageDetails", []):
            pushed = d.get("imagePushedAt")
            if isinstance(pushed, str):
                pushed = datetime.fromisoformat(pushed).timestamp()
            images[d["imageDigest"]] = {
                "tags": d.get("imageTags") or [],
                "pushed": float(pushed or 0),
                "size": d.get("imageSizeInBytes") or 0,
                "media_type": d.get("imageManifestMediaType") or "",
            }
        token = page.get("nextToken")
        if not token:
            return images


def registry_token(alias, repository):
    """A bearer token the registry accepts for the repository: the account's, else an anonymous one."""
    def candidates():
        try:
            yield aws("ecr-public", "get-authorization-token")["authorizationData"]["authorizationToken"]
        except (subprocess.CalledProcessError, KeyError):
            pass
        with urllib.request.urlopen(REGISTRY + "/token/", timeout=30) as r:
            yield json.load(r)["token"]

    url = f"{REGISTRY}/v2/{alias}/{repository}/tags/list?n=1"
    for token in candidates():
        try:
            request = urllib.request.Request(url, headers={"Authorization": "Bearer " + token})
            with urllib.request.urlopen(request, timeout=30):
                return token
        except urllib.error.URLError:
            continue
    raise RuntimeError(f"no token the registry accepts for {alias}/{repository}")


def fetch_manifest(alias, repository, digest, token):
    """The manifest as a dict, or None when the registry would not give it."""
    url = f"{REGISTRY}/v2/{alias}/{repository}/manifests/{digest}"
    request = urllib.request.Request(url, headers={"Authorization": "Bearer " + token, "Accept": ACCEPT})
    for attempt in range(5):
        try:
            with urllib.request.urlopen(request, timeout=60) as r:
                return json.load(r)
        except urllib.error.HTTPError as e:
            if e.code not in (429, 500, 502, 503, 504):
                return None
        except (urllib.error.URLError, TimeoutError, ValueError):
            pass
        time.sleep(2 ** attempt)
    return None


def references(manifest):
    """(children, subject) of a manifest: what an index lists and what it is attached to."""
    children = [m["digest"] for m in manifest.get("manifests") or [] if "digest" in m]
    subject = (manifest.get("subject") or {}).get("digest")
    return children, subject


def plan(images, manifests, cutoff):
    """The digests to keep and to delete.

    images: list_images(); manifests: digest -> fetched manifest or None, for
    every image index and every manifest without a version tag; cutoff:
    manifests pushed after this time are kept.
    """
    def digest_tags_only(image):
        return bool(image["tags"]) and all(DIGEST_TAG.match(t) for t in image["tags"])

    keep = set()
    for digest, image in images.items():
        tagged = image["tags"] and not digest_tags_only(image)
        unreadable = digest in manifests and manifests[digest] is None
        if tagged or image["pushed"] > cutoff or unreadable:
            keep.add(digest)

    children, subject, named = {}, {}, {}
    for digest, manifest in manifests.items():
        if manifest:
            children[digest], subject[digest] = references(manifest)
    for digest, image in images.items():
        if digest_tags_only(image):
            named[digest] = {"sha256:" + DIGEST_TAG.match(t).group(1) for t in image["tags"]}

    changed = True
    while changed:
        changed = False
        for digest in images:
            if digest in keep:
                continue
            if subject.get(digest) in keep or named.get(digest, set()) & keep:
                keep.add(digest)
                changed = True
        for digest in list(keep):
            for child in children.get(digest, []):
                if child in images and child not in keep:
                    keep.add(child)
                    changed = True
    return keep, set(images) - keep


def delete(repository, images, digests):
    """Deletes the digests, image indexes before what they list. Returns the failures."""
    order = sorted(digests, key=lambda d: images[d]["media_type"] not in INDEX_TYPES)
    failures = []
    for i in range(0, len(order), 100):
        ids = [{"imageDigest": d} for d in order[i:i + 100]]
        result = aws("ecr-public", "batch-delete-image", "--repository-name", repository,
                     "--image-ids", json.dumps(ids))
        failures += result.get("failures", [])
    return failures


def list_repositories():
    names, token = [], None
    while True:
        args = ["ecr-public", "describe-repositories", "--max-results", "1000"]
        if token:
            args += ["--next-token", token]
        page = aws(*args)
        names += [r["repositoryName"] for r in page.get("repositories", [])]
        token = page.get("nextToken")
        if not token:
            return names


def largest(count):
    """The repositories of the registry with the largest sum of manifest sizes."""
    def sizes(repository):
        images = list_images(repository).values()
        untagged = [i for i in images if not i["tags"]]
        return {
            "repository": repository,
            "manifests": len(images),
            "size": sum(i["size"] for i in images),
            "untagged": len(untagged),
            "untagged_size": sum(i["size"] for i in untagged),
        }

    with ThreadPoolExecutor(max_workers=4) as pool:
        rows = list(pool.map(sizes, list_repositories()))
    rows.sort(key=lambda r: r["size"], reverse=True)
    print(f"The {count} largest of {len(rows)} repositories, {gb(sum(r['size'] for r in rows))} GB in all:")
    for r in rows[:count]:
        print(f"  {r['repository']}: {gb(r['size'])} GB in {r['manifests']} manifests, "
              f"{gb(r['untagged_size'])} GB in {r['untagged']} untagged")
    return rows[:count], rows


def gb(size):
    return f"{size / 1e9:.1f}"


def collect(alias, repository, grace_days, do_delete):
    images = list_images(repository)
    token = registry_token(alias, repository)
    wanted = [d for d, i in images.items()
              if i["media_type"] in INDEX_TYPES or not i["tags"] or all(DIGEST_TAG.match(t) for t in i["tags"])]
    with ThreadPoolExecutor(max_workers=8) as pool:
        bodies = pool.map(lambda d: fetch_manifest(alias, repository, d, token), wanted)
    manifests = dict(zip(wanted, bodies))
    cutoff = (datetime.now(timezone.utc) - timedelta(days=grace_days)).timestamp()
    keep, garbage = plan(images, manifests, cutoff)

    unreadable = sum(1 for m in manifests.values() if m is None)
    row = {
        "repository": repository,
        "manifests": len(images),
        "kept": len(keep),
        "garbage": len(garbage),
        "garbage_gb": gb(sum(images[d]["size"] for d in garbage)),
        "unreadable": unreadable,
        "deleted": 0,
        "failures": 0,
    }
    print(f"{repository}: {len(images)} manifests, {len(keep)} kept, {len(garbage)} to delete "
          f"({row['garbage_gb']} GB of manifest sizes), {unreadable} unreadable and kept")
    for digest in sorted(garbage, key=lambda d: images[d]["pushed"]):
        image = images[digest]
        pushed = datetime.fromtimestamp(image["pushed"], timezone.utc).strftime("%Y-%m-%d")
        print(f"  {'delete' if do_delete else 'would delete'} {digest} pushed {pushed} "
              f"{image['size']} bytes {' '.join(image['tags'])}".rstrip())
    if do_delete and garbage:
        failures = delete(repository, images, garbage)
        for f in failures:
            print(f"  FAILED {f.get('imageId', {}).get('imageDigest')}: {f.get('failureCode')} {f.get('failureReason')}")
        row["failures"] = len(failures)
        row["deleted"] = len(garbage) - len(failures)
    return row


def main():
    parser = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    parser.add_argument("repositories", nargs="+", metavar="REPOSITORY")
    parser.add_argument("--alias", default="entigolabs", help="registry alias (default: entigolabs)")
    parser.add_argument("--grace-days", type=float, default=7,
                        help="keep everything pushed within this many days (default: 7)")
    parser.add_argument("--delete", action="store_true", help="delete; without it only report")
    parser.add_argument("--largest", type=int, default=0, metavar="N",
                        help="also report the N largest repositories of the registry")
    args = parser.parse_args()

    rows, failed = [], False
    for repository in args.repositories:
        try:
            rows.append(collect(args.alias, repository, args.grace_days, args.delete))
        except subprocess.CalledProcessError as e:
            print(f"{repository}: {e.stderr.strip() or e}", file=sys.stderr)
            failed = True
        except (RuntimeError, OSError) as e:
            print(f"{repository}: {e}", file=sys.stderr)
            failed = True
    failed = failed or any(r["failures"] for r in rows)

    top, every = [], []
    if args.largest > 0:
        try:
            top, every = largest(args.largest)
        except subprocess.CalledProcessError as e:
            print(f"largest repositories: {e.stderr.strip() or e}", file=sys.stderr)
            failed = True

    summary = os.environ.get("GITHUB_STEP_SUMMARY")
    if summary and top:
        with open(summary, "a") as f:
            f.write(f"### The {len(top)} largest of {len(every)} repositories\n\n")
            f.write("| Repository | GB | Manifests | Untagged GB | Untagged manifests |\n|---|---|---|---|---|\n")
            for r in top:
                f.write(f"| {r['repository']} | {gb(r['size'])} | {r['manifests']} | "
                        f"{gb(r['untagged_size'])} | {r['untagged']} |\n")
            f.write(f"\nAll repositories: {gb(sum(r['size'] for r in every))} GB. Sizes add up manifest sizes, "
                    "which count a layer once per image that uses it, so they rank repositories "
                    "and are more than what is billed.\n\n")
    if summary and rows:
        with open(summary, "a") as f:
            f.write(f"### ECR Public garbage collection ({'delete' if args.delete else 'dry run'}, "
                    f"grace {args.grace_days:g} days)\n\n")
            f.write("| Repository | Manifests | Kept | Garbage | Garbage GB | Deleted | Failed | Unreadable |\n")
            f.write("|---|---|---|---|---|---|---|---|\n")
            for r in rows:
                f.write(f"| {r['repository']} | {r['manifests']} | {r['kept']} | {r['garbage']} | "
                        f"{r['garbage_gb']} | {r['deleted']} | {r['failures']} | {r['unreadable']} |\n")
            f.write("\nGarbage GB adds up manifest sizes, which count a layer once per image that uses it.\n")
    sys.exit(1 if failed else 0)


if __name__ == "__main__":
    main()
