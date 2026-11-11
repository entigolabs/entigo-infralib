#!/bin/bash
# The EKS node AMI release versions pinned in modules/aws/eks/main.tf against
# the ones SSM recommends for the module's default Kubernetes version. Run by
# entigo-infralib-test's scripts/report.sh after its own aws checks (the
# Report workflow, or ../entigo-infralib-test/scripts/report.sh aws by hand);
# needs AWS credentials, which the workflow passes on. The output lines have
# the shape the update prompt expects: "EKS <type>_* newer version X, current Y".
set -euo pipefail
cd "$(dirname "$0")/.."

version=$(awk '/variable "eks_cluster_version"/,/^}/' modules/aws/eks/variables.tf | grep default | cut -d'"' -f2)
region=${AWS_REGION:-eu-north-1}

check() { # <label> <current> <ssm parameter>
  local latest
  latest=$(aws ssm get-parameter --name "$3" --region "$region" --query Parameter.Value --output text 2>/dev/null) \
    || { echo "$1 recommended release version not found in SSM $3"; return; }
  [ "$2" = "$latest" ] || echo "$1 newer version $latest, current $2 (modules/aws/eks/main.tf)"
}

check "EKS AL2023_*" "$(grep '"AL2023_x86_64_STANDARD"' modules/aws/eks/main.tf | cut -d'"' -f4)" \
  "/aws/service/eks/optimized-ami/$version/amazon-linux-2023/x86_64/standard/recommended/release_version"
check "EKS BOTTLEROCKET_*" "$(grep '"BOTTLEROCKET_x86_64"' modules/aws/eks/main.tf | cut -d'"' -f4)" \
  "/aws/service/bottlerocket/aws-k8s-$version/x86_64/latest/image_version"
