#!/bin/bash
# Google Cloud-specific functions

export PROVIDER="google"

# Get working directory for this environment
get_work_dir() {
    if [ "$LOCAL_MODE" == "true" ]; then
        echo "/tmp/project"
    else
        echo "/project"
    fi
}

# Copy project files from bucket to local directory
# Usage: copy_from_bucket <bucket> <source_path> <dest_path>
copy_from_bucket() {
    local bucket="$1"
    local source_path="$2"
    local dest_path="$3"
    local exclude=()

    if [ "$TERRAFORM_CACHE" != "true" ]; then
        exclude=(--exclude '.terraform/*')
    fi

    shackle cp "gs://${bucket}/${source_path}" "${dest_path}" --recursive --quiet "${exclude[@]}"
}

# Copy file to bucket
# Usage: copy_to_bucket <local_file> <bucket> <dest_path>
copy_to_bucket() {
    local local_file="$1"
    local bucket="$2"
    local dest_path="$3"

    shackle cp "${local_file}" "gs://${bucket}/${dest_path}" --quiet
}

# Sync terraform cache to bucket
sync_terraform_cache() {
    local bucket="$1"
    local prefix="$2"

    echo "Syncing .terraform back to bucket"
    shackle sync .terraform "gs://${bucket}/steps/${prefix}/.terraform" --delete --quiet
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
        shackle cp "gs://${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz" /project/tf.tar.gz --quiet
        if [ $? -ne 0 ]; then
            echo "Unable to find artifacts from plan stage! gs://${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz"
            exit 4
        fi
        cd /project/ && tar -xzf tf.tar.gz
        cd "steps/$TF_VAR_prefix"
    fi
}

# Upload plan artifact after plan stage
upload_plan_artifact() {
    if [ "$LOCAL_MODE" != "true" ]; then
      cd ../..
      tar -czf tf.tar.gz "steps/$TF_VAR_prefix"
      echo "Copy plan to Google Storage"
      shackle cp tf.tar.gz "gs://${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz" --quiet
      upload_plan_json
    fi
}

# Artifact Registry login: an access token with a fixed user name
GAR_TOKEN_USERNAME="oauth2accesstoken"
get_gar_token() {
    shackle registry-token google --region "$GOOGLE_REGION"
}

# Get Kubernetes credentials
get_k8s_credentials() {
    shackle kubeconfig google --cluster "$KUBERNETES_CLUSTER_NAME" --region "$GOOGLE_REGION" --project "$GOOGLE_PROJECT"
}
