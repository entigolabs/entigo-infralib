#!/bin/bash
# Oracle Cloud-specific functions
#
# The step files, plans and outputs are copied with shackle through the native
# Object Storage API, with the identity of the job (a resource principal in a
# container, the OCI configuration file locally).
#
# The S3-compatible API is used by the OpenTofu backend alone, which has no
# native Oracle backend. Its endpoint (AWS_ENDPOINT_URL_S3), region (AWS_REGION)
# and Customer Secret Key credentials (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY)
# are supplied via the environment by the agent (oracle.Resources.GetBackendEnv).
# The settings below keep that backend away from the operator's ~/.aws/config
# during local runs.

export PROVIDER="oracle"

export AWS_CONFIG_FILE="/tmp/oracle-aws-config"

# The AWS SDK adds CRC32 integrity checksums to uploads by default
# (x-amz-sdk-checksum-algorithm / x-amz-checksum-*), which OCI's S3-compatible API
# rejects ("x-amz-content-sha256 must be UNSIGNED-PAYLOAD or a valid sha256").
# GETs are unaffected; PUTs fail. Only send checksums when the operation requires it.
export AWS_REQUEST_CHECKSUM_CALCULATION=when_required
export AWS_RESPONSE_CHECKSUM_VALIDATION=when_required

# Object Storage location of a bucket path: oci://<bucket>@<namespace>/<path>.
# The namespace is the first label of the S3-compatible endpoint
# (https://<namespace>.compat.objectstorage.<region>.oraclecloud.com), unless
# OCI_NAMESPACE names it.
oci_location() {
    local namespace="$OCI_NAMESPACE"
    if [ -z "$namespace" ]; then
        namespace="${AWS_ENDPOINT_URL_S3#*://}"
        namespace="${namespace%%.*}"
    fi
    if [ -z "$namespace" ]; then
        echo "Unable to tell the Object Storage namespace: set OCI_NAMESPACE or AWS_ENDPOINT_URL_S3" >&2
        return 1
    fi
    echo "oci://${1}@${namespace}/${2}"
}

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
    local exclude=()

    if [ "$TERRAFORM_CACHE" != "true" ]; then
        exclude=(--exclude '*.terraform/*')
    fi

    # The Customer Secret Key is for the OpenTofu backend; an ArgoCD step has
    # no backend and runs without one.
    case "$COMMAND" in
        argocd-*) ;;
        *)
            if [ -z "$AWS_ACCESS_KEY_ID" ]; then
                echo "AWS_ACCESS_KEY_ID is not set — the Oracle s3-compatible backend needs a Customer Secret Key."
                echo "The agent provisions one automatically under session-token or API-key auth; a resource-principal"
                echo "(in-container) run reuses the persisted key, so bootstrap the deployment locally once first."
                exit 1
            fi
            ;;
    esac

    local source
    source=$(oci_location "$bucket" "$source_path") || exit 1
    shackle cp "$source" "${dest_path}" --recursive --quiet "${exclude[@]}" || exit 1
}

# Copy file to bucket
# Usage: copy_to_bucket <local_file> <bucket> <dest_path>
copy_to_bucket() {
    local local_file="$1"
    local bucket="$2"
    local dest_path="$3"

    local destination
    destination=$(oci_location "$bucket" "$dest_path") || exit 1
    shackle cp "${local_file}" "$destination" --quiet || exit 1
}

# Sync terraform cache to bucket
sync_terraform_cache() {
    local bucket="$1"
    local prefix="$2"

    echo "Syncing .terraform back to bucket"
    local destination
    destination=$(oci_location "$bucket" "steps/${prefix}/.terraform") || exit 1
    shackle sync .terraform "$destination" --delete --quiet
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
        shackle cp "$(oci_location "$INFRALIB_BUCKET" "${TF_VAR_prefix}-tf.tar.gz")" /tmp/project/tf.tar.gz --quiet
        if [ $? -ne 0 ]; then
            echo "Unable to find artifacts from plan stage! ${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz"
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
      echo "Copy plan to Oracle Object Storage"
      copy_to_bucket tf.tar.gz "$INFRALIB_BUCKET" "${TF_VAR_prefix}-tf.tar.gz"
      upload_plan_json
    fi
}

# Get Kubernetes credentials for an OKE cluster.
# KUBERNETES_CLUSTER_NAME may be the cluster OCID directly, or a display name resolved
# via OCI_COMPARTMENT_ID. The identity is a resource principal in a container
# (signalled by the variable the OCI runtime injects), otherwise the OCI
# configuration file; shackle picks it, and the tokens the kubeconfig asks it
# for later use the same one.
get_k8s_credentials() {
    if [[ "$KUBERNETES_CLUSTER_NAME" != ocid1.cluster.* ]] && [ -z "$OCI_COMPARTMENT_ID" ]; then
        echo "OCI_COMPARTMENT_ID must be set to resolve OKE cluster '$KUBERNETES_CLUSTER_NAME' by name"
        exit 1
    fi
    # Default to the private endpoint: OKE clusters are private-only in most
    # setups, and the in-container (Container Instances, same VCN) execution model
    # can reach it. Override with OCI_KUBE_ENDPOINT=PUBLIC_ENDPOINT when needed.
    local endpoint=(--private-endpoint)
    if [ "$OCI_KUBE_ENDPOINT" == "PUBLIC_ENDPOINT" ]; then
        endpoint=()
    fi
    shackle kubeconfig oracle --cluster "$KUBERNETES_CLUSTER_NAME" --region "$OCI_REGION" \
        ${OCI_COMPARTMENT_ID:+--compartment "$OCI_COMPARTMENT_ID"} "${endpoint[@]}" || exit 1
}
