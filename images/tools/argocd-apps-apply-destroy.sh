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

if [ ! -f "${app_file}.sync-destroy" ] #In plan stage we mark the apps that are not synced so we would only sync the ones we need to sync.
then
  echo "Skip $app_name"
  exit 0
fi

echo "Deleting $app_name"
if yq -r '.spec.sources[0].path' $app_file | grep -q "modules/k8s/external-dns"
then
  echo "Waiting 70 seconds for external DNS to complete last round of sync before being deleted."
  sleep 70
fi

if yq -r '.spec.sources[0].path' $app_file | grep -vEq "modules/k8s/argocd"
then
  kubectl patch applications.argoproj.io $app_name -n ${ARGOCD_NAMESPACE} -p '{"metadata": {"finalizers": ["resources-finalizer.argocd.argoproj.io"]}}' --type merge
  kubectl delete --grace-period=600 applications.argoproj.io $app_name -n ${ARGOCD_NAMESPACE}
  kubectl delete --grace-period=600 ns $app_name
else  
  echo "For ArgoCD we only remove Ingress"
  kubectl delete --grace-period=600 applications.argoproj.io $app_name -n ${ARGOCD_NAMESPACE}
  kubectl delete ingress -n ${ARGOCD_NAMESPACE} --all
fi


echo "###############"
