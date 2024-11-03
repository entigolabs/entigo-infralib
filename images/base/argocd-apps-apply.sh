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
app_name=`yq -r '.metadata.name' $app_file`

if [ "$app_name" == "" ]
then
  echo "Unable to find .metadata.name in $app_file."
  exit 27
fi

#We want to use argocd for sync but in first runs it is not yet available so we fall back to auto sync in Applications.
if [ "$ARGOCD_AUTH_TOKEN" != "" ]
then
  if [ ! -f "${app_file}.sync" ] #In plan stage we mark the apps that are not synced so we would only sync the ones we need to sync.
  then
    echo "Skip $app_name"
    exit 0
  fi

  argocd --server ${ARGOCD_HOSTNAME} --grpc-web app sync --prune $app_name
  if [ $? -ne 0 ]
  then
    echo "Failed $app_name sync"
    exit 24
  fi
  argocd --server ${ARGOCD_HOSTNAME} --grpc-web app wait --timeout 600 --health --sync --operation $app_name
  if [ $? -ne 0 ]
  then
    echo "Failed $app_name wait"
    exit 25
  fi
else #Fall back to Application auto sync when we can not get argo token.
  echo "AutoSync $app_name"
  kubectl patch -n ${ARGOCD_NAMESPACE} app $app_name --type merge --patch '{"spec": {"syncPolicy": {"automated": {"selfHeal": true}}}}'
fi

echo "###############"
