#!/bin/bash

echo "Containers without requests or limits."
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH/../..

kubectl get pods -A -o json | jq -r '.items[] | . as $pod | .spec.containers[] | select((.resources.requests == null or .resources.limits == null) and (.name | IN("aws-node", "aws-eks-nodeagent", "kube-proxy") | not)) | "\($pod.metadata.namespace)/\($pod.metadata.name)/\(.name)"'
