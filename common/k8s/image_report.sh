#!/bin/bash
SCRIPTPATH="$(
  cd -- "$(dirname "$0")" >/dev/null 2>&1
  pwd -P
)"
cd $SCRIPTPATH/../..

gcloud container clusters get-credentials pri-infra-gke --region $GOOGLE_REGION >/dev/null 2>&1
GOOGLE_PRI_CONTAINERS=$(kubectl get pods -A -o json | jq -r .items[].spec.containers[].image | sort | uniq)
gcloud container clusters get-credentials biz-infra-gke --region $GOOGLE_REGION >/dev/null 2>&1
GOOGLE_BIZ_CONTAINERS=$(kubectl get pods -A -o json | jq -r .items[].spec.containers[].image | sort | uniq)
# aws eks update-kubeconfig --region $AWS_REGION --name pri-infra-eks > /dev/null
# AWS_PRI_CONTAINERS=`kubectl get pods -A -o json | jq -r .items[].spec.containers[].image | sort | uniq`
# aws eks update-kubeconfig --region $AWS_REGION --name biz-infra-eks > /dev/null
# AWS_BIZ_CONTAINERS=`kubectl get pods -A -o json | jq -r .items[].spec.containers[].image | sed 's/biz-net-ecr-proxy/pri-net-ecr-proxy/g' | sort | uniq`

# aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin 602401143452.dkr.ecr.$AWS_REGION.amazonaws.com > /dev/null 2>&1
# aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin 877483565445.dkr.ecr.$AWS_REGION.amazonaws.com > /dev/null 2>&1

total=0
security=0
registry=0
for image in $(echo "877483565445.dkr.ecr.eu-north-1.amazonaws.com/pri-net-ecr-proxy-hub/entigolabs/entigo-infralib-testing:latest 877483565445.dkr.ecr.eu-north-1.amazonaws.com/pri-net-ecr-proxy-hub/entigolabs/entigo-infralib-agent:latest $AWS_BIZ_CONTAINERS $AWS_PRI_CONTAINERS $GOOGLE_PRI_CONTAINERS $GOOGLE_BIZ_CONTAINERS" | tr ' ' '\n' | sort | uniq | tr '\n' ' '); do
  # let total++
  # docker pull $image >/dev/null 2>&1
  # VULN=$(trivy image -q -f json --severity CRITICAL $image 2>/dev/null | jq -r 'select(.Results != null) | .Results[] | select(.Vulnerabilities != null) | .Vulnerabilities[] | select(.Severity != null) | .Severity' | sort | uniq -c | tr "\n" " ")
  # if [ $? -ne 0 ]; then
  #   echo "$image Scan failed."
  # else
  #   if [ "$VULN" != "" ]; then
  #     echo "$image $VULN"
  #     let security++
  #   fi
  # fi
  if [[ ! $image =~ (^877483565445\.dkr\.ecr\.$AWS_REGION\.amazonaws\.com|^602401143452\.dkr\.ecr\.$AWS_REGION\.amazonaws\.com|^europe-north1-artifactregistry\.gcr\.io|^oci\.external-secrets\.io|^xpkg\.upbound\.io) ]]; then
    echo "$image does not use Internal Registry"
    let registry++
  fi
done

echo "Total images $total, critical security issues $security, proxy registry issues $registry"
