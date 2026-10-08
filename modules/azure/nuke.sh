#!/bin/bash
# Removes the Entigo Infralib test environments (biz, pri, spoke) from the Azure test subscription.
#
# Only resource groups tagged by infralib are deleted: the agent resource group <prefix>-infralib-<location>
# (created-by=entigo-infralib-agent) and leftovers named <prefix>-*-<location> tagged created-by=entigo-infralib
# (AKS node and Container Apps environment resource groups, normally deleted with their cluster/environment).
# Also deleted: the prefixes' NS delegations in the persistent parent zone. Never touched: the parent zone resource
# group, NetworkWatcherRG and anything else in the subscription.
#
# Key Vaults are only soft-deleted with their resource group: the next build recovers the agent vault with its secrets
# (acr-proxy credentials) and the fixed-name kms vault. kms keys are soft-deleted first, so the recovered vault has no
# live keys and recover_soft_deleted_keys brings them back instead of "already exists".
#
# DRY_RUN=1 only lists what would be deleted.
PREFIXES="${PREFIXES:-biz pri spoke}"
LOCATION="${AZURE_LOCATION:-westeurope}"
PARENT_ZONE_RG="${PARENT_ZONE_RG:-infralib-permanent-$LOCATION}"
PARENT_ZONE="${PARENT_ZONE:-azure.infralib.entigo.io}"
WAIT_MINUTES="${WAIT_MINUTES:-60}"

if [ "$AZURE_SUBSCRIPTION_ID" == "" ]
then
  echo "ERROR: AZURE_SUBSCRIPTION_ID must be set, the script never guesses the subscription."
  exit 1
fi

if [ "$AZURE_CLIENT_SECRET" != "" ]
then
  export AZURE_CONFIG_DIR="${AZURE_CONFIG_DIR:-${RUNNER_TEMP:-$HOME}/.azure-infralib-nuke}"
  az login --service-principal -u "$AZURE_CLIENT_ID" -p "$AZURE_CLIENT_SECRET" --tenant "$AZURE_TENANT_ID" -o none || exit 1
fi
az account set -s "$AZURE_SUBSCRIPTION_ID" || exit 1
if [ "$(az account show --query id -o tsv)" != "$AZURE_SUBSCRIPTION_ID" ]
then
  echo "ERROR: active subscription is not $AZURE_SUBSCRIPTION_ID"
  exit 1
fi

run() {
  if [ "$DRY_RUN" == "1" ]
  then
    echo "  DRY_RUN: $*"
  else
    "$@"
  fi
}

echo "Nuking Azure test environments"
echo "  subscription: $AZURE_SUBSCRIPTION_ID"
echo "  location:     $LOCATION"
echo "  prefixes:     $PREFIXES"
echo "  parent zone:  $PARENT_ZONE_RG/$PARENT_ZONE"

for prefix in $PREFIXES
do
  # NS delegations live outside the resource groups; a dangling delegation could be taken over by another tenant
  for ns in $(az network dns record-set ns list -g "$PARENT_ZONE_RG" -z "$PARENT_ZONE" \
      --query "[?metadata.Prefix && starts_with(metadata.Prefix, '$prefix-') && metadata.\"created-by\"=='entigo-infralib'].name" -o tsv 2>/dev/null)
  do
    echo "deleting NS delegation $ns.$PARENT_ZONE"
    run az network dns record-set ns delete -g "$PARENT_ZONE_RG" -z "$PARENT_ZONE" -n "$ns" --yes
  done

  rg="$prefix-infralib-$LOCATION"
  if [ "$(az group show -n "$rg" --query 'tags."created-by"' -o tsv 2>/dev/null)" != "entigo-infralib-agent" ]
  then
    echo "$rg: not found or not created by the agent, skipped"
    continue
  fi

  for vault in $(az keyvault list -g "$rg" --query "[?tags.\"created-by\"=='entigo-infralib'].name" -o tsv)
  do
    for key in $(az keyvault key list --vault-name "$vault" --query "[].name" -o tsv)
    do
      echo "soft-deleting key $vault/$key"
      run az keyvault key delete --vault-name "$vault" -n "$key" -o none
    done
  done

  for lock in $(az lock list -g "$rg" --query "[].id" -o tsv)
  do
    echo "deleting lock $lock"
    run az lock delete --ids "$lock"
  done

  echo "deleting $rg"
  run az group delete -n "$rg" --yes --no-wait
done

[ "$DRY_RUN" == "1" ] && exit 0

leftovers() {
  for prefix in $PREFIXES
  do
    az group list --query "[?(name=='$prefix-infralib-$LOCATION' && tags.\"created-by\"=='entigo-infralib-agent') || (starts_with(name, '$prefix-') && ends_with(name, '-$LOCATION') && tags.\"created-by\"=='entigo-infralib')].name" -o tsv
  done
}

echo "waiting up to $WAIT_MINUTES minutes for the resource groups to be deleted"
end=$(( $(date +%s) + WAIT_MINUTES * 60 ))
while [ "$(date +%s)" -lt "$end" ]
do
  remaining=$(leftovers)
  [ -z "$remaining" ] && break
  # Azure deletes managed resource groups with their cluster/environment; orphans whose owner is gone stay
  for rg in $remaining
  do
    [ "$(az group show -n "$rg" --query properties.provisioningState -o tsv 2>/dev/null)" == "Succeeded" ] || continue
    owner=$(az group show -n "$rg" --query managedBy -o tsv 2>/dev/null)
    if [ -z "$owner" ] || ! az resource show --ids "$owner" -o none 2>/dev/null
    then
      echo "deleting leftover $rg"
      az group delete -n "$rg" --yes --no-wait
    fi
  done
  sleep 60
done

remaining=$(leftovers)
if [ -n "$remaining" ]
then
  echo "ERROR: still present after $WAIT_MINUTES minutes:"
  echo "$remaining"
  exit 1
fi
echo "done"
