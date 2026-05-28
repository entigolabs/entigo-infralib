#!/usr/bin/env python3
# Usage:
#   Helm:    python3 common/sbom.py helm <chart_dir> <chart_name> <version>
#   OpenTofu: python3 common/sbom.py tofu <module_dir> <module_name> <version> <module_type>

import json
import re
import subprocess
import sys
import uuid
from datetime import datetime, timezone


def make_document(name, version, namespace_suffix):
    # Base SPDX 2.3 document structure
    return {
        "spdxVersion": "SPDX-2.3",
        "dataLicense": "CC0-1.0",
        "SPDXID": "SPDXRef-DOCUMENT",
        "name": f"{name}-{version}",
        "documentNamespace": f"https://entigolabs.io/sbom/{namespace_suffix}/{version}/{uuid.uuid4()}",
        "creationInfo": {
            "licenseListVersion": "3.28",
            "creators": ["Organization: Entigolabs", "Tool: entigo-infralib-sbom"],
            "created": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
        },
        "packages": [],
        "relationships": []
    }


def add_package(doc, spdx_id, name, version, download_location, package_type, purl):
    doc["packages"].append({
        "name": name,
        "SPDXID": spdx_id,
        "versionInfo": version,
        "supplier": "NOASSERTION",
        "downloadLocation": download_location,
        "filesAnalyzed": False,
        "licenseConcluded": "NOASSERTION",
        "licenseDeclared": "NOASSERTION",
        "copyrightText": "NOASSERTION",
        "primaryPackagePurpose": package_type,
        "externalRefs": [
            {
                "referenceCategory": "PACKAGE-MANAGER",
                "referenceType": "purl",
                "referenceLocator": purl
            }
        ]
    })
    doc["relationships"].append({
        "spdxElementId": "SPDXRef-DOCUMENT",
        "relatedSpdxElement": spdx_id,
        "relationshipType": "DESCRIBES"
    })


def safe_spdx_id(name):
    # SPDX IDs only allow letters, numbers and dashes
    return re.sub(r"[^a-zA-Z0-9\-]", "-", name)


def generate_helm_sbom(chart_dir, chart_name, version):
    doc = make_document(chart_name, version, f"k8s/{chart_name}")

    # Parse Chart.yaml for subchart dependencies
    chart_yaml_path = f"{chart_dir}/Chart.yaml"
    with open(chart_yaml_path) as f:
        content = f.read()

    # Extract dependencies block entries
    dep_blocks = re.findall(
        r"- .*?(?=\n- |\Z)",
        re.search(r"dependencies:(.*?)(?=\n\w|\Z)", content, re.DOTALL).group(1) if "dependencies:" in content else "",
        re.DOTALL
    )

    for block in dep_blocks:
        name_match    = re.search(r"name:\s*(\S+)", block)
        version_match = re.search(r"version:\s*(\S+)", block)
        repo_match    = re.search(r"repository:\s*(\S+)", block)
        alias_match   = re.search(r"alias:\s*(\S+)", block)

        if not name_match or not version_match:
            continue

        dep_name    = name_match.group(1)
        dep_version = version_match.group(1)
        dep_repo    = repo_match.group(1) if repo_match else "NOASSERTION"
        dep_alias   = alias_match.group(1) if alias_match else dep_name
        spdx_id     = f"SPDXRef-helm-{safe_spdx_id(dep_alias)}"

        add_package(
            doc, spdx_id, dep_name, dep_version,
            download_location=dep_repo,
            package_type="APPLICATION",
            purl=f"pkg:helm/{dep_name}@{dep_version}"
        )

    # Run helm template to extract container images
    try:
        result = subprocess.run(
            ["helm", "template", chart_dir],
            capture_output=True, text=True, check=True
        )
        images = set(
            re.findall(r"image:\s*['\"]?([^\s'\"]+)['\"]?", result.stdout)
        )
        for image in sorted(images):
            # Parse image into name and tag
            tag     = "latest"
            img_ref = image
            if ":" in image.split("/")[-1]:
                img_ref, tag = image.rsplit(":", 1)
            spdx_id = f"SPDXRef-image-{safe_spdx_id(img_ref.replace('/', '-'))}"
            add_package(
                doc, spdx_id, img_ref, tag,
                download_location=f"https://{img_ref}" if "/" in img_ref else f"https://hub.docker.com/_/{img_ref}",
                package_type="CONTAINER",
                purl=f"pkg:oci/{img_ref}@{tag}"
            )
    except subprocess.CalledProcessError as e:
        print(f"Warning: helm template failed: {e.stderr}", file=sys.stderr)

    return doc


def generate_tofu_sbom(module_dir, module_name, version, module_type):
    doc = make_document(module_name, version, f"{module_type}/{module_name}")

    versions_tf = f"{module_dir}/versions.tf"
    try:
        with open(versions_tf) as f:
            content = f.read()
    except FileNotFoundError:
        print(f"Warning: versions.tf not found in {module_dir}", file=sys.stderr)
        return doc

    # Parse required_providers blocks
    providers = re.findall(
        r'(\w+)\s*=\s*\{[^}]*source\s*=\s*"([^"]+)"[^}]*version\s*=\s*"([^"]+)"',
        content
    )

    for alias, source, prov_version in providers:
        spdx_id = f"SPDXRef-provider-{safe_spdx_id(alias)}"
        add_package(
            doc, spdx_id, source, prov_version,
            download_location=f"https://registry.opentofu.org/providers/{source}",
            package_type="LIBRARY",
            purl=f"pkg:terraform/{source}@{prov_version}"
        )

    # Parse required_version for OpenTofu itself
    tofu_version_match = re.search(r'required_version\s*=\s*"([^"]+)"', content)
    if tofu_version_match:
        add_package(
            doc, "SPDXRef-opentofu", "opentofu", tofu_version_match.group(1),
            download_location="https://opentofu.org",
            package_type="APPLICATION",
            purl=f"pkg:generic/opentofu@{tofu_version_match.group(1)}"
        )

    return doc


def main():
    if len(sys.argv) < 5:
        print(f"Usage: {sys.argv[0]} helm <chart_dir> <chart_name> <version>", file=sys.stderr)
        print(f"       {sys.argv[0]} tofu <module_dir> <module_name> <version> <module_type>", file=sys.stderr)
        sys.exit(1)

    mode = sys.argv[1]

    if mode == "helm":
        chart_dir, chart_name, version = sys.argv[2], sys.argv[3], sys.argv[4]
        doc = generate_helm_sbom(chart_dir, chart_name, version)
    elif mode == "tofu":
        module_dir, module_name, version, module_type = sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5]
        doc = generate_tofu_sbom(module_dir, module_name, version, module_type)
    else:
        print(f"Unknown mode: {mode}", file=sys.stderr)
        sys.exit(1)

    print(json.dumps(doc, indent=2))


if __name__ == "__main__":
    main()
