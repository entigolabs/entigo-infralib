#!/bin/bash
echo "#################################"
echo "Containers without requests or limits."
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH/../..

kubectl get pods -A -o json | jq -r '.items[] | . as $pod | .spec.containers[] | select((.resources.requests == null or .resources.limits == null) and (.name | IN("default-http-backend","prometheus-to-sd-exporter","dnsmasq","sidecar","event-exporter","autoscaler" ,"netd", "aws-node", "aws-eks-nodeagent", "kube-proxy", "calico-node", "ip-masq-agent", "calico-typha", "cilium-agent") | not)) | "\($pod.metadata.namespace)/\($pod.metadata.name)/\(.name)"'

echo "#################################"
echo "Namespaces without PSA"

printf "%-40s %-12s %-12s\n" "NAMESPACE" "ENFORCE" "WARN"
printf "%-40s %-12s %-12s\n" "---------" "-------" "----"

kubectl get namespaces -o json | jq -r '.items[] |
  select(.metadata.name != "kube-system") |
  select(.metadata.name != "kube-public") |
  select(.metadata.name != "default") |
  select(.metadata.name != "kube-node-lease") |
  select(.metadata.name != "biz") |
  select(.metadata.name != "pri") |
  select(.metadata.name != "prometheus-pri") |
  select(.metadata.name != "prometheus-biz") |
  select(.metadata.name != "wireguard-pri") |
  select(.metadata.name != "alloy-pri") |
  select(.metadata.name != "alloy-biz") |
  select(.metadata.name != "gke-managed-system") |
  select(.metadata.name != "gke-managed-volumepopulator") |
  (.metadata.labels["pod-security.kubernetes.io/enforce"] // "privileged") as $enforce |
  (.metadata.labels["pod-security.kubernetes.io/warn"] // "privileged") as $warn |
  select($enforce != "restricted" or $warn != "restricted") |
  [.metadata.name, $enforce, $warn] | @tsv
' | while IFS=$'\t' read -r ns enforce warn; do
  printf "%-40s %-12s %-12s\n" "$ns" "$enforce" "$warn"
done
