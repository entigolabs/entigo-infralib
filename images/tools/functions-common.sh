#!/bin/bash
# Common functions shared across all cloud providers

# TF_TOOL controls which binary is used: terraform (default) or tofu
TF_TOOL=${TF_TOOL:-terraform}

# Run Go tests
run_tests() {
    if [ ! -f go.mod ]; then
        cd /common && go mod download -x && cd /app
        go mod init github.com/entigolabs/entigo-infralib
        go mod edit -require github.com/entigolabs/entigo-infralib-common@v0.0.0 -replace github.com/entigolabs/entigo-infralib-common=/common
        go mod tidy
    fi
    cd test && go test -timeout $ENTIGO_INFRALIB_TEST_TIMEOUT
    exit $?
}

# Prepare working directory and copy files for plan stage
prepare_plan_stage() {
    echo "Need to copy project files from bucket $INFRALIB_BUCKET"

    WORK_DIR=$(get_work_dir)
    if [ "$LOCAL_MODE" == "true" ]; then
        mkdir -p /tmp/plans/$TF_VAR_prefix/
    fi
    mkdir -p $WORK_DIR/steps/$TF_VAR_prefix
    cd $WORK_DIR

    copy_from_bucket "$INFRALIB_BUCKET" "steps/$TF_VAR_prefix" "./steps/$TF_VAR_prefix"

    if [ ! -d "steps/$TF_VAR_prefix" ]; then
        find .
        echo "Unable to find path steps/$TF_VAR_prefix"
        exit 5
    fi
    cd "steps/$TF_VAR_prefix"
}

# Prepare terraform environment
prepare_terraform() {
    git_login
    setup_ca_certificates
    terraform_init
}

# Run terraform plan
terraform_plan() {
    $TF_TOOL plan -no-color -out ${TF_VAR_prefix}.tf-plan -input=false
    if [ $? -ne 0 ]; then
        echo "Failed to create TF plan!"
        exit 6
    fi
    $TF_TOOL show -json ${TF_VAR_prefix}.tf-plan > ${TF_VAR_prefix}-plan.json
    if [ $? -ne 0 ]; then
        echo "Failed to create json plan from TF plan!"
        exit 6
    fi
}

# Run terraform plan for destroy
terraform_plan_destroy() {
    $TF_TOOL plan -destroy -no-color -out ${TF_VAR_prefix}.tf-plan-destroy -input=false
    if [ $? -ne 0 ]; then
        echo "Failed to create TF destroy plan!"
        exit 6
    fi
}

# Run terraform apply
terraform_apply() {
    if [ "$TERRAFORM_CACHE" == "true" ]; then
        sync_terraform_cache "$INFRALIB_BUCKET" "$TF_VAR_prefix"
    fi
    $TF_TOOL apply -no-color -input=false ${TF_VAR_prefix}.tf-plan
    if [ $? -ne 0 ]; then
        echo "Apply failed!"
        exit 11
    fi
    $TF_TOOL output -json > terraform-output.json
    if [ $? -ne 0 ]; then
        echo "Output failed!"
        exit 12
    fi
    copy_to_bucket "terraform-output.json" "$INFRALIB_BUCKET" "$TF_VAR_prefix/terraform-output.json"
}

# Run terraform apply for destroy
terraform_apply_destroy() {
    $TF_TOOL apply -no-color -input=false ${TF_VAR_prefix}.tf-plan-destroy
    if [ $? -ne 0 ]; then
        echo "Apply destroy failed!"
        exit 11
    fi
}

running_jobs() {
    local still_running=""
    local status_changed=0

    for p in $PIDS; do
        pid=$(echo $p | cut -d"=" -f1)
        name=$(echo $p | cut -d"=" -f2)

        # Skip if we already know about this job's completion
        if [[ $COMPLETED == *$p* ]] || [[ $FAIL == *$p* ]]; then
            continue
        fi

        if kill -0 $pid 2>/dev/null; then
            still_running="$still_running\n~ $(basename $name) Running"
        else
            # Check if job completed successfully
            if wait $pid 2>/dev/null; then
                echo "✓ $(basename $name) Done"
                COMPLETED="$COMPLETED $p"
            else
                echo "✗ $(basename $name) Failed"
                FAIL="$FAIL $p"
            fi
            status_changed=1
        fi
    done

    # Only show running jobs if the list has changed
    if [ "$still_running" != "$LAST_RUNNING" ]; then
        if [ ! -z "$still_running" ]; then
            echo -e "-------$still_running"
        fi
        LAST_RUNNING="$still_running"
    fi
}

get_priority_from_yaml() {
    local yaml_file="$1"
    local priority

    # Extract priority from metadata.annotations."infralib.entigo.io/sync-wave" using yq
    priority=$(yq -r '.metadata.annotations."infralib.entigo.io/sync-wave" // 100' "$yaml_file" 2>/dev/null)

    # Ensure we have a numeric value
    if ! [[ "$priority" =~ ^[0-9]+$ ]]; then
        priority=100
    fi

    echo "$priority"
}

# Wait for all background jobs to complete
wait_for_jobs() {
    FAIL=""
    COMPLETED=""
    LAST_RUNNING=""
    while true; do
        sleep 2
        running_jobs
        total_done=$(echo "$COMPLETED $FAIL" | wc -w)
        total_jobs=$(echo "$PIDS" | wc -w)

        if [ $total_done -eq $total_jobs ]; then
            break
        fi
    done
}

# Print logs for completed and failed jobs
print_job_logs() {
    for p in $COMPLETED; do
        name=$(echo $p | cut -d"=" -f2)
        cat $name.log
    done

    for p in $FAIL; do
        name=$(echo $p | cut -d"=" -f2)
        cat $name.log
    done
}

# Print failed jobs and exit
handle_failed_jobs() {
    local exit_message="$1"
    local exit_code="${2:-21}"

    if [ ! -z "$FAIL" ]; then
        for p in $FAIL; do
            name=$(echo $p | cut -d"=" -f2)
            echo "#######################################"
            echo "ERROR LOG FOR $name"
            cat $name.log
        done

        echo "Failed jobs were:"
        for p in $FAIL; do
            echo "  - $(basename $(echo $p | cut -d"=" -f2))"
        done
        echo "$exit_message"
        exit $exit_code
    fi
}

# Authenticate git repos if available
git_login() {
    for var in "${!GIT_AUTH_SOURCE_@}"; do
        printf 'Configure credentials for %s=%s\n' "$var" "${!var}"
        PASSWORD="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_PASSWORD/g')"
        USERNAME="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_USERNAME/g')"
        if [[ "${!var}" =~ ^(https?|ssh):// || "${!var}" =~ ^git@ ]]; then
            # GIT source, rewrite the host to include credentials
            SOURCE="$(echo ${!var} | sed -n 's|.*://\([^/]*\).*|\1|p')"
            git config --global url."https://${!USERNAME}:${!PASSWORD}@${SOURCE}".insteadOf https://${SOURCE}
        else
            # OCI source, OpenTofu reads oci_credentials blocks from the CLI config
            # The label supports a registry domain with an optional repository path prefix
            SOURCE="${!var#oci://}"
            echo "oci_credentials \"${SOURCE}\" {
  username = \"${!USERNAME}\"
  password = \"${!PASSWORD}\"
}" >> $HOME/.tofurc
        fi
    done

    # Only create ~/.tofurc if AWS_REGION is set
    if [ -n "$AWS_REGION" ]; then
      # Get current account number
      ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

      cat >> "$HOME/.tofurc" <<EOF
oci_credentials "${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com" {
  docker_credentials_helper = "ecr-login"
}
EOF
    elif [ ! -z "$GOOGLE_REGION" ]; then
      cat >> "$HOME/.tofurc" <<EOF
oci_credentials "${GOOGLE_REGION}-docker.pkg.dev" {
  docker_credentials_helper = "gcloud"
}
EOF
    fi
}

# Setup CA certificates
setup_ca_certificates() {
  if [ -d "ca-certificates" ]
  then
    echo "Additional CA certificates found, running update-ca-certificates"
    cp ca-certificates/*.crt /usr/local/share/ca-certificates/
    sudo update-ca-certificates
  fi
}

# Initialize terraform with backend config
terraform_init() {
    cat backend.conf
    if [ $? -ne 0 ]; then
        echo "Unable to find backend.conf file"
        exit 100
    fi
    #
    find .terraform/providers -name 'terraform-provider-*' -exec chmod +x {} +

    # tofu requires -reconfigure when taking over a Terraform-initialized backend
    $TF_TOOL init -input=false -reconfigure -backend-config=backend.conf
    if [ $? -ne 0 ]; then
        echo "Terraform init failed."
        exit 14
    fi
}
