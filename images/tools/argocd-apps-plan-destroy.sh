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
else
  app_namespace=$ARGOCD_NAMESPACE
fi

if kubectl get applications.argoproj.io $app_name -n $app_namespace -o jsonpath='{.metadata.name}' >/dev/null 2>&1; then
    APP_EXISTED="yes"
    
    if yq -r '.spec.sources[0].path' $app_file | grep -vEq "modules/k8s/crossplane-core|modules/k8s/crossplane-aws|modules/k8s/crossplane-google|modules/k8s/aws-storageclass"
    then
      kubectl patch -n $app_namespace applications.argoproj.io $app_name --type=json -p="[{'op': 'remove', 'path': '/spec/syncPolicy/automated'}]" > /dev/null 2>&1
      touch $app_file.sync-destroy
    fi
else
    APP_EXISTED="no"
fi

echo "###############"
