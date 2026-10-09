#!/bin/bash
# Azure-specific functions
#
# INFRALIB_BUCKET is the storage account name. All files live in the "tfstate"
# blob container, same as the agent (azure/blob.go parseFilePath).

export PROVIDER="azure"
[ -z "$AZURE_SUBSCRIPTION_ID" ] && echo "AZURE_SUBSCRIPTION_ID must be set" && exit 1
export ARM_SUBSCRIPTION_ID="$AZURE_SUBSCRIPTION_ID"
if [ -z "$ARM_CLIENT_ID" ] && [ -n "$AZURE_CLIENT_SECRET" ]; then
    export ARM_CLIENT_ID="$AZURE_CLIENT_ID" ARM_CLIENT_SECRET="$AZURE_CLIENT_SECRET" ARM_TENANT_ID="$AZURE_TENANT_ID"
fi
AZ_CONTAINER="tfstate"

azure_login() {
    if az account show >/dev/null 2>&1; then
        :
    elif [ -n "$ARM_CLIENT_ID" ] && [ -n "$ARM_CLIENT_SECRET" ]; then
        az login --service-principal -u "$ARM_CLIENT_ID" -p "$ARM_CLIENT_SECRET" --tenant "$ARM_TENANT_ID" >/dev/null || exit 1
    elif [ -n "$IDENTITY_ENDPOINT" ] || [ "$AZURE_CONTAINER_APP_JOB" == "true" ]; then
        az login --identity ${AZURE_CLIENT_ID:+--client-id "$AZURE_CLIENT_ID"} >/dev/null || exit 1
        export ARM_USE_MSI=false ARM_USE_CLI=true
        unset ARM_TENANT_ID
    else
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
    local tmp_dir=$(mktemp -d)

    az storage blob download-batch --auth-mode login --account-name "$bucket" -s "$AZ_CONTAINER" \
        --pattern "${source_path}/*" -d "$tmp_dir" --overwrite --no-progress >/dev/null || exit 1
    mkdir -p "$dest_path"
    cp -a "$tmp_dir/$source_path/." "$dest_path/"
    rm -rf "$tmp_dir"

    if [ "$TERRAFORM_CACHE" != "true" ]; then
        rm -rf "$dest_path/.terraform"
    fi
}

copy_to_bucket() {
    local local_file="$1"
    local bucket="$2"
    local dest_path="$3"

    az storage blob upload --auth-mode login --account-name "$bucket" -c "$AZ_CONTAINER" \
        -n "$dest_path" -f "$local_file" --overwrite --no-progress >/dev/null || exit 1
}

sync_terraform_cache() {
    local bucket="$1"
    local prefix="$2"

    echo "Syncing .terraform back to bucket"
    az storage blob delete-batch --auth-mode login --account-name "$bucket" -s "$AZ_CONTAINER" \
        --pattern "steps/${prefix}/.terraform/*" >/dev/null
    az storage blob upload-batch --auth-mode login --account-name "$bucket" -d "$AZ_CONTAINER" \
        --destination-path "steps/${prefix}/.terraform" -s .terraform --overwrite --no-progress >/dev/null
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
        az storage blob download --auth-mode login --account-name "$INFRALIB_BUCKET" -c "$AZ_CONTAINER" \
            -n "${TF_VAR_prefix}-tf.tar.gz" -f /tmp/project/tf.tar.gz --overwrite --no-progress >/dev/null
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

get_acr_login_server() {
    if [ -n "$AZURE_ACR_NAME" ]; then
        az acr show -n "$AZURE_ACR_NAME" --query loginServer -o tsv
        return
    fi
    local servers
    servers=$(az acr list -g "$AZURE_RESOURCE_GROUP" --query '[].loginServer' -o tsv) || return 1
    if [ $(echo "$servers" | grep -c .) -gt 1 ]; then
        echo "Several container registries in $AZURE_RESOURCE_GROUP, set AZURE_ACR_NAME: $(echo $servers)" >&2
        return 1
    fi
    echo "$servers"
}

ACR_TOKEN_USERNAME="00000000-0000-0000-0000-000000000000"
get_acr_token() {
    az acr login -n "${1%%.*}" --expose-token --query accessToken -o tsv
}

get_k8s_credentials() {
    az aks get-credentials -g "$AZURE_RESOURCE_GROUP" -n "$KUBERNETES_CLUSTER_NAME" --overwrite-existing || exit 1
    kubelogin convert-kubeconfig -l azurecli || exit 1
    echo "Kubeconfig set for AKS cluster $KUBERNETES_CLUSTER_NAME in $AZURE_RESOURCE_GROUP"
}
