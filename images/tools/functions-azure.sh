#!/bin/bash
# Azure-specific functions
#
# INFRALIB_BUCKET is the storage account name. All files live in the "tfstate"
# blob container, same as the agent (azure/blob.go parseFilePath).

export PROVIDER="azure"
export ARM_SUBSCRIPTION_ID="$AZURE_SUBSCRIPTION_ID"
AZ_CONTAINER="tfstate"

# Login order: existing session (mounted ~/.azure), service principal, managed identity
azure_login() {
    if az account show >/dev/null 2>&1; then
        :
    elif [ -n "$ARM_CLIENT_ID" ] && [ -n "$ARM_CLIENT_SECRET" ]; then
        az login --service-principal -u "$ARM_CLIENT_ID" -p "$ARM_CLIENT_SECRET" --tenant "$ARM_TENANT_ID" >/dev/null || exit 1
    elif [ -n "$IDENTITY_ENDPOINT" ] || [ "$AZURE_CONTAINER_APP_JOB" == "true" ]; then
        az login --identity ${AZURE_CLIENT_ID:+--client-id "$AZURE_CLIENT_ID"} >/dev/null || exit 1
    else
        echo "No Azure credentials found: mount ~/.azure, set ARM_CLIENT_ID/ARM_CLIENT_SECRET/ARM_TENANT_ID or run with a managed identity"
        exit 1
    fi
    az account set --subscription "$AZURE_SUBSCRIPTION_ID" || exit 1
}
azure_login

# Get working directory for this environment
get_work_dir() {
    echo "/tmp/project"
}

# Copy project files from bucket to local directory
# Usage: copy_from_bucket <bucket> <source_path> <dest_path>
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

    # download-batch has no exclude option
    if [ "$TERRAFORM_CACHE" != "true" ]; then
        rm -rf "$dest_path/.terraform"
    fi
}

# Copy file to bucket
# Usage: copy_to_bucket <local_file> <bucket> <dest_path>
copy_to_bucket() {
    local local_file="$1"
    local bucket="$2"
    local dest_path="$3"

    az storage blob upload --auth-mode login --account-name "$bucket" -c "$AZ_CONTAINER" \
        -n "$dest_path" -f "$local_file" --overwrite --no-progress >/dev/null || exit 1
}

# Sync terraform cache to bucket
sync_terraform_cache() {
    local bucket="$1"
    local prefix="$2"

    echo "Syncing .terraform back to bucket"
    az storage blob delete-batch --auth-mode login --account-name "$bucket" -s "$AZ_CONTAINER" \
        --pattern "steps/${prefix}/.terraform/*" >/dev/null
    az storage blob upload-batch --auth-mode login --account-name "$bucket" -d "$AZ_CONTAINER" \
        --destination-path "steps/${prefix}/.terraform" -s .terraform --overwrite --no-progress >/dev/null
}

# Fetch plan artifact for apply stage
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

# Upload plan artifact after plan stage
upload_plan_artifact() {
    if [ "$LOCAL_MODE" != "true" ]; then
      cd ../..
      tar -czf tf.tar.gz "steps/$TF_VAR_prefix"
      echo "Copy plan to Azure Blob Storage"
      copy_to_bucket tf.tar.gz "$INFRALIB_BUCKET" "${TF_VAR_prefix}-tf.tar.gz"
      upload_plan_json
    fi
}

# ACR login server: AZURE_ACR_NAME, otherwise the only registry in AZURE_RESOURCE_GROUP
# Prints nothing when there is no registry, fails when there are several and no override
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

# ACR refresh token (valid ~3h), used as password with username ACR_TOKEN_USERNAME
# Usage: get_acr_token <login server>
ACR_TOKEN_USERNAME="00000000-0000-0000-0000-000000000000"
get_acr_token() {
    az acr login -n "${1%%.*}" --expose-token --query accessToken -o tsv
}

# Get Kubernetes credentials for an AKS cluster, kubelogin reuses the az session
get_k8s_credentials() {
    az aks get-credentials -g "$AZURE_RESOURCE_GROUP" -n "$KUBERNETES_CLUSTER_NAME" --overwrite-existing || exit 1
    kubelogin convert-kubeconfig -l azurecli || exit 1
    echo "Kubeconfig set for AKS cluster $KUBERNETES_CLUSTER_NAME in $AZURE_RESOURCE_GROUP"
}
