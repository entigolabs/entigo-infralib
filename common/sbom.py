#!/usr/bin/env python3
# Usage:
#   Helm:     python3 common/sbom.py helm <chart_dir> <chart_name> <version>
#   OpenTofu: python3 common/sbom.py tofu <module_dir> <module_name> <version> <module_type>

import json
import re
import subprocess
import sys
import os
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


def add_root_package(doc, spdx_id, name, version, download_location, package_type, purl):
    # Root package — DOCUMENT DESCRIBES root
    doc["packages"].append({
        "name": name,
        "SPDXID": spdx_id,
        "versionInfo": version,
        "supplier": "Organization: Entigolabs",
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


def add_dependency(doc, root_spdx_id, spdx_id, name, version, download_location,
                   package_type, purl, comment=None, relationship="DEPENDS_ON"):
    # Dependency package
    pkg = {
        "name": name,
        "SPDXID": spdx_id,
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
    }
    # Omit versionInfo if not known — "latest" or empty string defeats SBOM purpose
    if version:
        pkg["versionInfo"] = version
    if comment:
        pkg["comment"] = comment
    doc["packages"].append(pkg)

    # BUILD_TOOL_OF direction is reversed: tool BUILD_TOOL_OF module
    if relationship == "BUILD_TOOL_OF":
        doc["relationships"].append({
            "spdxElementId": spdx_id,
            "relatedSpdxElement": root_spdx_id,
            "relationshipType": "BUILD_TOOL_OF"
        })
    else:
        doc["relationships"].append({
            "spdxElementId": root_spdx_id,
            "relatedSpdxElement": spdx_id,
            "relationshipType": relationship
        })


def safe_spdx_id(name):
    # SPDX IDs only allow letters, numbers and dashes
    return re.sub(r"[^a-zA-Z0-9\-]", "-", name)


def generate_helm_sbom(chart_dir, chart_name, version):
    doc = make_document(chart_name, version, f"k8s/{chart_name}")

    root_id  = "SPDXRef-chart-root"
    repo_url = f"ghcr.io/entigolabs/entigo-infralib-release/k8s/{chart_name}"
    add_root_package(
        doc, root_id, chart_name, version,
        download_location=f"oci://{repo_url}",
        package_type="INSTALL",
        # Digest not available at SBOM generation time — tag only
        purl=f"pkg:oci/{chart_name}?repository_url={repo_url}&tag={version}"
    )

    # Parse Chart.yaml for subchart dependencies
    with open(f"{chart_dir}/Chart.yaml") as f:
        content = f.read()

    dep_blocks = re.findall(
        r"- .*?(?=\n- |\Z)",
        re.search(r"dependencies:(.*?)(?=\n\w|\Z)", content, re.DOTALL).group(1) if "dependencies:" in content else "",
        re.DOTALL
    )

    for block in dep_blocks:
        name_match    = re.search(r"name:\s*(\S+)", block)
        version_match = re.search(r"version:\s*(\S+)", block)
        repo_match    = re.search(r"repository:\s*(\S+)", block)

        if not name_match or not version_match:
            continue

        dep_name    = name_match.group(1)
        dep_version = version_match.group(1)
        dep_repo    = repo_match.group(1) if repo_match else "NOASSERTION"
        spdx_id     = f"SPDXRef-helm-{safe_spdx_id(dep_name)}"

        add_dependency(
            doc, root_id, spdx_id, dep_name, dep_version,
            download_location=dep_repo,
            package_type="INSTALL",
            purl=f"pkg:generic/{dep_name}@{dep_version}?download_url={dep_repo}"
        )

    # Run helm template to extract container images
    try:
        cmd = ["helm", "template", chart_dir]
        # Some static tests require test/static_values.yaml to render;
        # include it when present (the test folder still exists at SBOM time)
        static_values = f"{chart_dir}/test/static_values.yaml"
        if os.path.isfile(static_values):
            cmd += ["-f", static_values]
        result = subprocess.run(
            cmd,
            capture_output=True, text=True, check=True
        )
        images = set(re.findall(r"image:\s*['\"]?([^\s'\"]+)['\"]?", result.stdout))
        for image in sorted(images):
            tag     = "latest"
            img_ref = image
            if ":" in image.split("/")[-1]:
                img_ref, tag = image.rsplit(":", 1)
            spdx_id = f"SPDXRef-image-{safe_spdx_id(img_ref.replace('/', '-'))}"

            parts = img_ref.split("/")
            if parts[0] in ("docker.io", "index.docker.io"):
                if len(parts) > 2 and parts[1] == "library":
                    img_url = f"https://hub.docker.com/_/{parts[2]}"
                else:
                    img_url = f"https://hub.docker.com/r/{'/'.join(parts[1:])}"
            elif "/" not in img_ref:
                img_url = f"https://hub.docker.com/_/{img_ref}"
            else:
                img_url = f"https://{img_ref}"

            add_dependency(
                doc, root_id, spdx_id, img_ref, tag,
                download_location=img_url,
                package_type="CONTAINER",
                purl=f"pkg:oci/{img_ref.split('/')[-1]}?repository_url={img_ref}&tag={tag}"
            )
    except subprocess.CalledProcessError as e:
        print(f"Warning: helm template failed: {e.stderr}", file=sys.stderr)

    return doc


def generate_tofu_sbom(module_dir, module_name, version, module_type):
    doc = make_document(module_name, version, f"{module_type}/{module_name}")

    root_id  = "SPDXRef-module-root"
    repo_url = f"ghcr.io/entigolabs/entigo-infralib-release/{module_type}/{module_name}"
    add_root_package(
        doc, root_id, module_name, version,
        download_location=f"oci://{repo_url}",
        package_type="LIBRARY",
        purl=f"pkg:oci/{module_name}?repository_url={repo_url}&tag={version}"
    )

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
        namespace, prov_name = source.split("/") if "/" in source else ("", source)
        spdx_id = f"SPDXRef-provider-{safe_spdx_id(alias)}"
        # Fold namespace into name: hashicorp/aws -> hashicorp-aws
        purl_name = f"{namespace}-{prov_name}" if namespace else prov_name
        add_dependency(
            doc, root_id, spdx_id, source, prov_version,
            download_location=f"https://registry.terraform.io/providers/{source}/{prov_version}",
            package_type="LIBRARY",
            purl=f"pkg:generic/{purl_name}@{prov_version}?download_url=https://registry.terraform.io/providers/{source}/{prov_version}"
        )

    # OpenTofu is a build tool, not a shipped dependency
    # versionInfo omitted — required_version is a constraint not a pinned version
    tofu_version_match = re.search(r'required_version\s*=\s*"([^"]+)"', content)
    if tofu_version_match:
        constraint = tofu_version_match.group(1)
        add_dependency(
            doc, root_id, "SPDXRef-opentofu", "opentofu", None,
            download_location="https://opentofu.org",
            package_type="APPLICATION",
            purl="pkg:generic/opentofu",
            comment=f"Required version constraint: {constraint}",
            relationship="BUILD_TOOL_OF"
        )

    return doc


def generate_agent_sbom(entries_path, version):
    # entries_path: JSONL file, one module record per line, written by the
    # helm/tofu publish steps. Each module becomes a CONTAINS dependency,
    # referenced by an oci purl per registry with the package digest as checksum.
    doc = make_document("agent", version, "agent")

    root_id  = "SPDXRef-agent-root"
    repo_url = "ghcr.io/entigolabs/entigo-infralib-release/agent"
    add_root_package(
        doc, root_id, "agent", version,
        download_location=f"oci://{repo_url}",
        package_type="LIBRARY",
        # Digest not available for the agent package itself at generation time
        purl=f"pkg:oci/agent?repository_url={repo_url}&tag={version}"
    )

    with open(entries_path) as f:
        entries = [json.loads(line) for line in f if line.strip()]

    for e in entries:
        spdx_id = f"SPDXRef-module-{safe_spdx_id(e['type'])}-{safe_spdx_id(e['name'])}"

        # One oci purl per registry — same artifact, different manifest digest
        ext_refs = []
        checksums = []
        for reg in ("ghcr", "ecr"):
            r = e["registries"][reg]
            ext_refs.append({
                "referenceCategory": "PACKAGE-MANAGER",
                "referenceType": "purl",
                # Digest is the version component; tag kept as a qualifier
                "referenceLocator": f"pkg:oci/{e['name'].lower()}@{r['package']}"
                                    f"?repository_url={r['repository']}&tag={e['tag']}"
            })
            # Record both the package and its SBOM digest as checksums
            checksums.append({
                "algorithm": "SHA256",
                "checksumValue": r["package"].split(":", 1)[1]  # strip "sha256:"
            })

        doc["packages"].append({
            "name": f"{e['type']}/{e['name']}",
            "SPDXID": spdx_id,
            "versionInfo": e["tag"],
            "supplier": "Organization: Entigolabs",
            "downloadLocation": f"oci://{e['registries']['ghcr']['repository']}",
            "filesAnalyzed": False,
            "licenseConcluded": "NOASSERTION",
            "licenseDeclared": "NOASSERTION",
            "copyrightText": "NOASSERTION",
            "primaryPackagePurpose": "LIBRARY",
            "checksums": checksums,
            "externalRefs": ext_refs
        })
        doc["relationships"].append({
            "spdxElementId": root_id,
            "relatedSpdxElement": spdx_id,
            "relationshipType": "CONTAINS"
        })

    return doc

def main():
    if len(sys.argv) < 4:
        print(f"Usage: {sys.argv[0]} helm <chart_dir> <chart_name> <version>", file=sys.stderr)
        print(f"       {sys.argv[0]} tofu <module_dir> <module_name> <version> <module_type>", file=sys.stderr)
        print(f"       {sys.argv[0]} agent <entries_jsonl> <version>", file=sys.stderr)
        sys.exit(1)

    mode = sys.argv[1]

    if mode == "helm":
        chart_dir, chart_name, version = sys.argv[2], sys.argv[3], sys.argv[4]
        doc = generate_helm_sbom(chart_dir, chart_name, version)
    elif mode == "tofu":
        module_dir, module_name, version, module_type = sys.argv[2], sys.argv[3], sys.argv[4], sys.argv[5]
        doc = generate_tofu_sbom(module_dir, module_name, version, module_type)
    elif mode == "agent":
        entries_path, version = sys.argv[2], sys.argv[3]
        doc = generate_agent_sbom(entries_path, version)
    else:
        print(f"Unknown mode: {mode}", file=sys.stderr)
        sys.exit(1)

    print(json.dumps(doc, indent=2))


if __name__ == "__main__":
    main()
