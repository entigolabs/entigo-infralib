#!/bin/bash
# Azure-specific functions
#
# INFRALIB_BUCKET is the storage account name. All files live in the "tfstate"
# blob container, same as the agent (azure/blob.go parseFilePath).

export PROVIDER="azure"
[ -z "$AZURE_SUBSCRIPTION_ID" ] && echo "AZURE_SUBSCRIPTION_ID must be set" && exit 1
export ARM_SUBSCRIPTION_ID="$AZURE_SUBSCRIPTION_ID"
# OpenTofu reads a service principal from ARM_*, shackle from AZURE_*: whichever
# set is given, both get it.
if [ -z "$ARM_CLIENT_ID" ] && [ -n "$AZURE_CLIENT_SECRET" ]; then
    export ARM_CLIENT_ID="$AZURE_CLIENT_ID" ARM_CLIENT_SECRET="$AZURE_CLIENT_SECRET" ARM_TENANT_ID="$AZURE_TENANT_ID"
elif [ -z "$AZURE_CLIENT_SECRET" ] && [ -n "$ARM_CLIENT_SECRET" ]; then
    export AZURE_CLIENT_ID="$ARM_CLIENT_ID" AZURE_CLIENT_SECRET="$ARM_CLIENT_SECRET" AZURE_TENANT_ID="$ARM_TENANT_ID"
fi
AZ_CONTAINER="tfstate"

# The bucket, the registry and the cluster are reached with shackle, which
# takes the identity straight from the environment: the service principal
# above, the managed identity of the job, or a mounted az login. The az CLI
# is logged in for one case only: OpenTofu under a managed identity, where the
# azurerm provider authenticates through the CLI.
azure_login() {
    if [ -n "$ARM_CLIENT_ID" ] && [ -n "$ARM_CLIENT_SECRET" ]; then
        return
    fi
    if [ -n "$IDENTITY_ENDPOINT" ] || [ "$AZURE_CONTAINER_APP_JOB" == "true" ]; then
        case "$COMMAND" in
            argocd-*) return ;;
        esac
        az login --identity ${AZURE_CLIENT_ID:+--client-id "$AZURE_CLIENT_ID"} >/dev/null || exit 1
        export ARM_USE_MSI=false ARM_USE_CLI=true
        unset ARM_TENANT_ID
    elif ! az account show >/dev/null 2>&1; then
        echo "No Azure credentials found: mount ~/.azure, set ARM_CLIENT_ID/ARM_CLIENT_SECRET/ARM_TENANT_ID or run with a managed identity"
        exit 1
    fi
    az account set --subscription "$AZURE_SUBSCRIPTION_ID" || exit 1
}
azure_login

get_work_dir() {
    echo "/tmp/project"
}

copy_from_bucket() {
    local bucket="$1"
    local source_path="$2"
    local dest_path="$3"
    local exclude=()

    if [ "$TERRAFORM_CACHE" != "true" ]; then
        exclude=(--exclude '.terraform/*')
    fi

    shackle cp "az://${bucket}/${AZ_CONTAINER}/${source_path}" "${dest_path}" --recursive --quiet "${exclude[@]}" || exit 1
}

copy_to_bucket() {
    local local_file="$1"
    local bucket="$2"
    local dest_path="$3"

    shackle cp "${local_file}" "az://${bucket}/${AZ_CONTAINER}/${dest_path}" --quiet || exit 1
}

sync_terraform_cache() {
    local bucket="$1"
    local prefix="$2"

    echo "Syncing .terraform back to bucket"
    shackle sync .terraform "az://${bucket}/${AZ_CONTAINER}/steps/${prefix}/.terraform" --delete --quiet || exit 1
}

fetch_plan_artifact() {
    if [ "$LOCAL_MODE" == "true" ]; then
        if [ ! -d /tmp/project/steps/$TF_VAR_prefix ]; then
            echo "Unable to find plan! /tmp/project/steps/$TF_VAR_prefix"
            exit 4
        fi
        cd "/tmp/project/steps/$TF_VAR_prefix"
    else
        mkdir -p /tmp/project
        shackle cp "az://${INFRALIB_BUCKET}/${AZ_CONTAINER}/${TF_VAR_prefix}-tf.tar.gz" /tmp/project/tf.tar.gz --quiet
        if [ $? -ne 0 ]; then
            echo "Unable to find artifacts from plan stage! ${INFRALIB_BUCKET}/${AZ_CONTAINER}/${TF_VAR_prefix}-tf.tar.gz"
            exit 4
        fi
        cd /tmp/project/ && tar -xzf tf.tar.gz
        cd "steps/$TF_VAR_prefix"
    fi
}

upload_plan_artifact() {
    if [ "$LOCAL_MODE" != "true" ]; then
      cd ../..
      tar -czf tf.tar.gz "steps/$TF_VAR_prefix"
      echo "Copy plan to Azure Blob Storage"
      copy_to_bucket tf.tar.gz "$INFRALIB_BUCKET" "${TF_VAR_prefix}-tf.tar.gz"
      upload_plan_json
    fi
}

# Login server of the environment's container registry: the one AZURE_ACR_NAME
# names, else the one registry of the resource group; empty when it has none.
get_acr_login_server() {
    if [ -n "$AZURE_ACR_NAME" ]; then
        shackle registries azure --registry "$AZURE_ACR_NAME"
        return
    fi
    local servers
    servers=$(shackle registries azure --resource-group "$AZURE_RESOURCE_GROUP") || return 1
    if [ $(echo "$servers" | grep -c .) -gt 1 ]; then
        echo "Several container registries in $AZURE_RESOURCE_GROUP, set AZURE_ACR_NAME: $(echo $servers)" >&2
        return 1
    fi
    echo "$servers"
}

ACR_TOKEN_USERNAME="00000000-0000-0000-0000-000000000000"
get_acr_token() {
    shackle registry-token azure --registry "$1"
}

get_k8s_credentials() {
    shackle kubeconfig azure --cluster "$KUBERNETES_CLUSTER_NAME" --resource-group "$AZURE_RESOURCE_GROUP" \
        --subscription "$AZURE_SUBSCRIPTION_ID" || exit 1
    echo "Kubeconfig set for AKS cluster $KUBERNETES_CLUSTER_NAME in $AZURE_RESOURCE_GROUP"
}
