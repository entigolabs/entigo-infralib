#!/bin/bash
# AWS-specific functions

export PROVIDER="aws"

# Get working directory for this environment
get_work_dir() {
    if [ "$LOCAL_MODE" == "true" ]; then
        echo "/tmp/project"
    else
        echo "$CODEBUILD_SRC_DIR"
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
        exclude=(--exclude '*.terraform/*')
    fi

    shackle cp "s3://${bucket}/${source_path}" "${dest_path}" --recursive --quiet "${exclude[@]}"
}

# Copy file to bucket
# Usage: copy_to_bucket <local_file> <bucket> <dest_path>
copy_to_bucket() {
    local local_file="$1"
    local bucket="$2"
    local dest_path="$3"

    shackle cp "${local_file}" "s3://${bucket}/${dest_path}" --quiet
}

# Sync terraform cache to bucket
sync_terraform_cache() {
    local bucket="$1"
    local prefix="$2"

    echo "Syncing .terraform back to bucket"
    shackle sync .terraform "s3://${bucket}/steps/${prefix}/.terraform" --delete --quiet
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
        if [ ! -f $CODEBUILD_SRC_DIR_Plan/tf.tar.gz ]; then
            echo "Unable to find artifacts from plan stage! $CODEBUILD_SRC_DIR_Plan/tf.tar.gz"
            exit 4
        fi
        tar -xzf $CODEBUILD_SRC_DIR_Plan/tf.tar.gz
        cd "steps/$TF_VAR_prefix"
    fi
}

# Upload plan artifact after plan stage
upload_plan_artifact() {
    if [ "$LOCAL_MODE" != "true" ]; then
      # AWS uses CodeBuild artifacts, no explicit upload needed
      # Just create the tar.gz for CodeBuild to pick up
      cd ../..
      tar -czf tf.tar.gz "steps/$TF_VAR_prefix"
      upload_plan_json
    fi
}

# Host of the account's ECR registry in the region
get_ecr_registry() {
    shackle registries aws --region "$AWS_REGION"
}

# Get Kubernetes credentials
get_k8s_credentials() {
    shackle kubeconfig aws --cluster "$KUBERNETES_CLUSTER_NAME" --region "$AWS_REGION"
}
