#!/bin/bash
# Oracle Cloud-specific functions
#
# Object Storage is accessed through its S3-compatible API with the standard
# aws cli. The endpoint (AWS_ENDPOINT_URL_S3), region (AWS_REGION) and Customer
# Secret Key credentials (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY) are supplied
# via the environment by the agent (oracle.Resources.GetBackendEnv). The compat
# endpoint embeds the namespace (https://<ns>.compat.objectstorage.<region>...),
# so path-style addressing is mandatory — virtual-hosted style would resolve the
# bucket as a subdomain and fail. We isolate that setting in a throwaway config
# file so the operator's ~/.aws/config is never touched during local runs.

export PROVIDER="oracle"

export AWS_CONFIG_FILE="/tmp/oracle-aws-config"
aws configure set default.s3.addressing_style path

# aws cli v2 (>= 2.23) adds CRC32 integrity checksums to uploads by default
# (x-amz-sdk-checksum-algorithm / x-amz-checksum-*), which OCI's S3-compatible API
# rejects ("x-amz-content-sha256 must be UNSIGNED-PAYLOAD or a valid sha256").
# GETs are unaffected; PUTs fail. Only send checksums when the operation requires it.
export AWS_REQUEST_CHECKSUM_CALCULATION=when_required
export AWS_RESPONSE_CHECKSUM_VALIDATION=when_required

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
    local exclude_pattern=""

    if [ "$TERRAFORM_CACHE" != "true" ]; then
        exclude_pattern="--exclude *.terraform/*"
    fi

    if [ -z "$AWS_ACCESS_KEY_ID" ]; then
        echo "AWS_ACCESS_KEY_ID is not set — the Oracle s3-compatible backend needs a Customer Secret Key."
        echo "The agent provisions one automatically under session-token or API-key auth; a resource-principal"
        echo "(in-container) run reuses the persisted key, so bootstrap the deployment locally once first."
        exit 1
    fi

    # Retry: a just-provisioned Customer Secret Key can still be missing from some
    # Object Storage backend hosts (per-host eventual consistency across the region),
    # so individual objects transiently fail SignatureDoesNotMatch even after the
    # agent's readiness poll passes — the poll and this copy hit different hosts.
    # Intermediate failures are expected and logged tersely; the full error is only
    # shown if every attempt is exhausted.
    local attempt output rc
    for ((attempt = 1; attempt <= 12; attempt++)); do
        output=$(aws s3 --endpoint-url "$AWS_ENDPOINT_URL_S3" cp "s3://${bucket}/${source_path}" "${dest_path}" --recursive --no-progress ${exclude_pattern} 2>&1)
        rc=$?
        if [ $rc -eq 0 ]; then
            return 0
        fi
        echo "copy_from_bucket: credentials not fully propagated across Object Storage yet, retrying (attempt ${attempt})"
        sleep 10
    done
    echo "copy_from_bucket failed after ${attempt} attempts for s3://${bucket}/${source_path}:"
    echo "$output"
    exit $rc
}

# Copy file to bucket
# Usage: copy_to_bucket <local_file> <bucket> <dest_path>
copy_to_bucket() {
    local local_file="$1"
    local bucket="$2"
    local dest_path="$3"

    local output
    output=$(aws s3 --endpoint-url "$AWS_ENDPOINT_URL_S3" cp "${local_file}" "s3://${bucket}/${dest_path}" --no-progress 2>&1)
    local rc=$?
    if [ $rc -ne 0 ]; then
        echo "copy_to_bucket failed (exit $rc) for ${local_file} -> s3://${bucket}/${dest_path}:"
        echo "$output"
        exit $rc
    fi
}

# Sync terraform cache to bucket
sync_terraform_cache() {
    local bucket="$1"
    local prefix="$2"

    echo "Syncing .terraform back to bucket"
    aws s3 --endpoint-url "$AWS_ENDPOINT_URL_S3" sync .terraform "s3://${bucket}/steps/${prefix}/.terraform" --no-progress --quiet --delete
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
        aws s3 --endpoint-url "$AWS_ENDPOINT_URL_S3" cp "s3://${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz" /tmp/project/tf.tar.gz --no-progress
        if [ $? -ne 0 ]; then
            echo "Unable to find artifacts from plan stage! s3://${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz"
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
      aws s3 --endpoint-url "$AWS_ENDPOINT_URL_S3" cp tf.tar.gz "s3://${INFRALIB_BUCKET}/${TF_VAR_prefix}-tf.tar.gz" --no-progress
    fi
}

# OCI CLI auth mode: resource principal in-container (signalled by the env var the
# OCI runtime injects), otherwise the default ~/.oci/config the cli reads locally.
oci_cli_auth() {
    if [ -n "$OCI_RESOURCE_PRINCIPAL_VERSION" ]; then
        echo "--auth resource_principal"
    fi
}

# Get Kubernetes credentials for an OKE cluster.
# NOTE: unvalidated against a live OKE cluster (ArgoCD on Oracle is not exercised yet).
# KUBERNETES_CLUSTER_NAME may be the cluster OCID directly, or a display name resolved
# via ORACLE_COMPARTMENT_ID. Requires the oci CLI in the image.
get_k8s_credentials() {
    local cluster_id="$KUBERNETES_CLUSTER_NAME"
    if [[ "$cluster_id" != ocid1.cluster.* ]]; then
        if [ -z "$ORACLE_COMPARTMENT_ID" ]; then
            echo "ORACLE_COMPARTMENT_ID must be set to resolve OKE cluster '$KUBERNETES_CLUSTER_NAME' by name"
            exit 1
        fi
        cluster_id=$(oci ce cluster list $(oci_cli_auth) \
            --compartment-id "$ORACLE_COMPARTMENT_ID" \
            --name "$KUBERNETES_CLUSTER_NAME" \
            --lifecycle-state ACTIVE \
            --region "$OCI_REGION" \
            --query 'data[0].id' --raw-output 2>/dev/null)
        if [ -z "$cluster_id" ] || [ "$cluster_id" = "null" ]; then
            echo "Unable to find an active OKE cluster named '$KUBERNETES_CLUSTER_NAME' in compartment $ORACLE_COMPARTMENT_ID"
            exit 1
        fi
    fi
    mkdir -p "$HOME/.kube"
    # Default to the private endpoint: OKE clusters are private-only in most
    # setups, and the in-container (Container Instances, same VCN) execution model
    # can reach it. Override with ORACLE_KUBE_ENDPOINT=PUBLIC_ENDPOINT when needed.
    oci ce cluster create-kubeconfig $(oci_cli_auth) \
        --cluster-id "$cluster_id" \
        --file "$HOME/.kube/config" \
        --region "$OCI_REGION" \
        --token-version 2.0.0 \
        --kube-endpoint "${ORACLE_KUBE_ENDPOINT:-PRIVATE_ENDPOINT}"
}

# Get ArgoCD hostname — cloud-agnostic ingress lookup, same as the AWS provider.
get_argocd_hostname() {
    kubectl get ingress -n ${ARGOCD_NAMESPACE} -l app.kubernetes.io/component=server -o jsonpath='{.items[*].spec.rules[*].host}'
}
