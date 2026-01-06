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
    local exclude_pattern=""

    if [ "$TERRAFORM_CACHE" != "true" ]; then
        exclude_pattern="-x \.terraform/.*"
    fi

    gsutil -m -q rsync -r ${exclude_pattern} "gs://${bucket}/${source_path}" "${dest_path}"
}

# Copy file to bucket
# Usage: copy_to_bucket <local_file> <bucket> <dest_path>
copy_to_bucket() {
    local local_file="$1"
    local bucket="$2"
    local dest_path="$3"

    gsutil -m -q cp "${local_file}" "gs://${bucket}/${dest_path}"
}

# Sync terraform cache to bucket
sync_terraform_cache() {
    local bucket="$1"
    local prefix="$2"

    echo "Syncing .terraform back to bucket"
    gsutil -m -q rsync -d -r .terraform "gs://${bucket}/steps/${prefix}/.terraform"
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
        gsutil -m -q cp "gs://${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz" /project/tf.tar.gz
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
      tar -czf tf.tar.gz "steps/$prefix"
      echo "Copy plan to Google Storage"
      gsutil -m -q cp tf.tar.gz "gs://${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz"
    fi
}

# Get Kubernetes credentials
get_k8s_credentials() {
    gcloud container clusters get-credentials $KUBERNETES_CLUSTER_NAME --region $GOOGLE_REGION --project $GOOGLE_PROJECT
}

# Get ArgoCD hostname
get_argocd_hostname() {
    kubectl get httproute -n ${ARGOCD_NAMESPACE} -o jsonpath='{.items[*].spec.hostnames[*]}'
}
