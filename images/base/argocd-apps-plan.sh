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

echo "Plan $app_name"

kubectl patch -n ${ARGOCD_NAMESPACE} app $app_name --type=json -p="[{'op': 'remove', 'path': '/spec/syncPolicy/automated'}]"
kubectl apply -n ${ARGOCD_NAMESPACE} -f $app_file
if [ $? -ne 0 ]
then
  echo "Failed to kubectl apply ArgoCD Application $app_name!"
  exit 24
fi

if [ "$ARGOCD_AUTH_TOKEN" != "" ]
then
  
  STATUS=`argocd --server ${ARGOCD_HOSTNAME} --grpc-web app get -o tree --refresh $app_name -o json | jq -r -e .status.sync.status`
  if [ $? -ne 0 ]
  then
    echo "Failed to refresh ArgoCD Application $app_name!"
    exit 25
  fi
  echo "Status $app_name $STATUS"
  if [ "$STATUS" != "Synced" ]
  then
    touch $app_file.sync
    argocd --server ${ARGOCD_HOSTNAME} --grpc-web app diff --exit-code=false $app_name
  fi
fi
echo "Finish $app_name"
echo "###############"
