#!/bin/bash
if [ "$ARGOCD_NAMESPACE" == "" ]
then
  echo "Unable to get ArgoCD namespace."
  exit 29
fi

if [ "$1" == "" ]
then
  echo "First parameters has to be ArgoCD Application file."
  exit 28
fi

app_file=$1
app_name=$(yq -r '.metadata.name // ""' "$app_file")
if [ -z "$app_name" ]
then
  echo "Unable to find .metadata.name in $app_file."
  exit 27
fi

app_namespace=$(yq -r '.metadata.namespace // ""' "$app_file")
if [ -n "$app_namespace" ]
then
  echo "Found .metadata.namespace in $app_file. Overriding ARGOCD_NAMESPACE to $app_namespace"
  # Create namespace if it doesn't exist
  if ! kubectl get namespace "$app_namespace" &>/dev/null; then
    echo "Namespace $app_namespace does not exist. Creating..."
    kubectl create namespace "$app_namespace"
  fi
else
  app_namespace=$ARGOCD_NAMESPACE
fi

#local development hack
targetRevision=`yq -r '.spec.sources[0].targetRevision' $app_file`
if [ "$targetRevision" == "local" ]
then
  echo "Detected local target"
  path=`yq -r '.spec.sources[0].path' $app_file`
  repo=`yq -r '.spec.sources[0].repoURL' $app_file`
  repopod=`kubectl get pod -n $ARGOCD_NAMESPACE -l app.kubernetes.io/component=repo-server -o jsonpath='{.items[0].metadata.name}'`
  kubectl exec -c repo-server -n $ARGOCD_NAMESPACE $repopod -- bash -c "mkdir -p /tmp/conf/modules/k8s && rm -rf /tmp/conf/$path"
  kubectl cp /conf/$path ${ARGOCD_NAMESPACE}/${repopod}:/tmp/conf/modules/k8s
  kubectl exec -it -c repo-server -n ${ARGOCD_NAMESPACE} $repopod -- bash -c "cd /tmp/conf/ && git init; git add . && git config user.email 'agent@entigo.com' && git config user.name 'agent' && git commit -a -m'updates'"
  yq -i 'del(.spec.sources[0].targetRevision)' $app_file
fi

if kubectl get applications.argoproj.io $app_name -n $app_namespace -o jsonpath='{.metadata.name}' >/dev/null 2>&1; then
    APP_EXISTED="yes"
    kubectl patch -n $app_namespace applications.argoproj.io $app_name --type=json -p="[{'op': 'remove', 'path': '/spec/syncPolicy/automated'}]" > /dev/null 2>&1
else
    APP_EXISTED="no"
fi

kubectl apply -n $app_namespace -f $app_file
if [ $? -ne 0 ]
then
  echo "Failed to kubectl apply ArgoCD Application $app_name!"
  exit 24
fi

if [ "$USE_ARGOCD_CLI" == "true" ]
then
  argocd --server ${ARGOCD_HOSTNAME} --grpc-web app get --refresh --app-namespace $app_namespace ${app_name} > /dev/null
  if [ $? -ne 0 ]
  then
    echo "Failed to refresh ArgoCD Application $app_name!"
    exit 25
  fi

  STATUS=$(argocd --server ${ARGOCD_HOSTNAME} --grpc-web app get --app-namespace $app_namespace $app_name -o json | jq -r '"Status:\(.status.sync.status) Missing:\(if .status.resources then (.status.resources | map(select(.status == "OutOfSync" and .health.status == "Missing" and (.hook == null or .hook == false))) | length) else 0 end) Changed:\(if .status.resources then (.status.resources | map(select(.status == "OutOfSync" and .health.status != "Missing" and .health.status != null and .requiresPruning != true and (.hook == null or .hook == false))) | length) else 0 end) RequiresPruning:\(if .status.resources then (.status.resources | map(select(.requiresPruning == true and (.hook == null or .hook == false))) | length) else 0 end)"')
  if [ $? -ne 0 ]
  then
    echo "Failed to get ArgoCD Application $app_name!"
    exit 25
  fi
else
  if [ $APP_EXISTED == "no" ]
  then
    retry_count=0
    while [ $retry_count -lt 50 ]; do
      APPSTATUS=$(kubectl get applications.argoproj.io -n $app_namespace $app_name -o json | jq -r '.status.sync.status // "null"')
      # Check if the result is not "null"
      if [ "$APPSTATUS" != "null" ] && [ -n "$APPSTATUS" ]; then
        break
      fi    
      echo "Waiting for ArgoCD to process $app_name application. Status is Null."
      sleep 6
      retry_count=$((retry_count + 1))
    done
  fi

  # Build API resources lookup map once: "kind.group" -> "resourcename.group"
  declare -A api_map
  while IFS= read -r line; do
      name=$(echo "$line" | awk '{print $1}')
      kind=$(echo "$line" | awk '{print $NF}')
      apiversion=$(echo "$line" | awk '{print $(NF-2)}')
      api_map["${kind}.${apiversion}"]="${name}"
  done < <(kubectl api-resources --no-headers)

  MISSING=0
  CHANGED=0

  resources=$(kubectl get applications.argoproj.io -n $app_namespace ${app_name} -o json | jq -r '
  .status.resources[]
  | select(.status == "OutOfSync" and .requiresPruning != true and (.hook == null or .hook == false))
  | [.group // "", .version, .kind, .namespace // "", .name] | @tsv
')
  if [ $? -ne 0 ]
  then
    echo "Failed to get ArgoCD Application missing and changed objects $app_name!"
    exit 26
  fi
  while IFS=$'\t' read -r group version kind namespace name; do
    [[ -z "$kind" ]] && continue

    # Build apiversion key to match api_map format
    if [[ -n "$group" ]]; then
        apiversion="${group}/${version}"
    else
        apiversion="${version}"
    fi
    resource_type="${api_map["${kind}.${apiversion}"]}"
    # Skip if resource type not found - consider it missing
    if [[ -z "$resource_type" ]]; then
        ((MISSING++))
        continue
    fi

    # Check if resource exists
    if [[ -n "$namespace" ]]; then
        kubectl get "${resource_type}.${group}" "${name}" -n "${namespace}" &>/dev/null
    else
        kubectl get "${resource_type}.${group}" "${name}" &>/dev/null
    fi

    if [[ $? -eq 0 ]]; then
        ((CHANGED++))
    else
        ((MISSING++))
    fi
  done <<< "$resources"
  # Get RequiresPruning count
  PRUNE=$(kubectl get applications.argoproj.io -n $app_namespace ${app_name} -o json | jq '
    [.status.resources[] | select(.requiresPruning == true and (.hook == null or .hook == false))] | length
  ')
  if [ $? -ne 0 ]
  then
    echo "Failed to get ArgoCD Application pruned objects $app_name!"
    exit 26
  fi

  # Get sync status
  SYNC_STATUS=$(kubectl get applications.argoproj.io -n $app_namespace ${app_name} -o jsonpath='{.status.sync.status}')
  if [ $? -ne 0 ]
  then
    echo "Failed to get ArgoCD Application $app_name!"
    exit 26
  fi

  STATUS=$(echo "Status:${SYNC_STATUS} Missing:${MISSING} Changed:${CHANGED} RequiresPruning:${PRUNE}")
fi

echo "Status $app_name $STATUS"
if [ "$STATUS" != "Status:Synced Missing:0 Changed:0 RequiresPruning:0" ]
then
  touch $app_file.sync
  if [ "$APP_EXISTED" == "yes" -a "$USE_ARGOCD_CLI" == "true" ]
  then
    argocd --server ${ARGOCD_HOSTNAME} --grpc-web app diff --exit-code=false --app-namespace $app_namespace $app_name
  fi
fi
echo "###############"
