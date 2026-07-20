#!/bin/bash
# ArgoCD-specific functions

# Initialize ArgoCD connection
# Sets ARGOCD_HOSTNAME, ARGOCD_AUTH_TOKEN, USE_ARGOCD_CLI
init_argocd_connection() {
    setup_ca_certificates
    echo "COMMAND $COMMAND, cluster $KUBERNETES_CLUSTER_NAME region ${GOOGLE_REGION:-$AWS_REGION}"

    get_k8s_credentials
    export ARGOCD_HOSTNAME=$(get_argocd_hostname)
    export ARGOCD_AUTH_TOKEN=$(kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-infralib-token -o jsonpath="{.data.token}" | base64 -d)
    export USE_ARGOCD_CLI="false"

    if [ "$ARGOCD_AUTH_TOKEN" != "" -a "$ARGOCD_HOSTNAME" != "" ]; then
        TES_CONNECTION=$(argocd --server ${ARGOCD_HOSTNAME} --http-retry-max 5 --grpc-web app list)
        if [ $? -eq 0 ]; then
            echo "Connected to ArgoCD successfully."
            export USE_ARGOCD_CLI="true"
        fi
    fi
}

# Create/update ArgoCD repository secrets from GIT_AUTH_SOURCE_*, GIT_AUTH_USERNAME_*, GIT_AUTH_PASSWORD_* env variables
# Detects GIT vs OCI sources by URL scheme: https://, http://, ssh:// or git@ means GIT, anything else is an OCI registry
# Usage: argocd_repositories [namespace], defaults to $ARGOCD_NAMESPACE
argocd_repositories() {
    local namespace="${1:-$ARGOCD_NAMESPACE}"

    for var in "${!GIT_AUTH_SOURCE_@}"; do
        local NAME="${var#GIT_AUTH_SOURCE_}"
        local SOURCE="${!var}"
        local PASSWORD="GIT_AUTH_PASSWORD_${NAME}"
        local USERNAME="GIT_AUTH_USERNAME_${NAME}"
        # Kubernetes secret names must be lowercase RFC 1123
        local secret_name="repo-$(echo $NAME | tr 'A-Z_' 'a-z-')"

        if [[ "$SOURCE" =~ ^(https?|ssh):// || "$SOURCE" =~ ^git@ ]]; then
            # GIT repository secret
            echo "Applying git repository secret $secret_name in namespace $namespace."
            echo "apiVersion: v1
kind: Secret
metadata:
  name: ${secret_name}
  namespace: ${namespace}
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: git
  name: ${NAME}
  url: ${SOURCE}.git
  username: \"${!USERNAME}\"
  password: \"${!PASSWORD}\"" | kubectl apply -f - || { echo "Failed to create repository secret $secret_name"; exit 24; }
        else
            # OCI helm registry secret, ArgoCD expects the url without the oci:// prefix
            local url="${SOURCE#oci://}"
            echo "Applying OCI repository secret $secret_name in namespace $namespace."
            echo "apiVersion: v1
kind: Secret
metadata:
  name: ${secret_name}
  namespace: ${namespace}
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: helm
  enableOCI: \"true\"
  name: ${NAME}
  url: ${url}
  username: \"${!USERNAME}\"
  password: \"${!PASSWORD}\"" | kubectl apply -f - || { echo "Failed to create repository secret $secret_name"; exit 24; }
        fi
    done

    # Seed a temporary ECR credential secret on AWS until External Secrets takes over
    # ECR tokens are valid for 12 hours, ESO keeps the secret refreshed afterwards.
    # Excluded for Oracle: the OCI s3-compatible backend also sets AWS_REGION, but
    # these AWS ECR/STS calls do not apply there.
    if [ -n "$AWS_REGION" ] && [ -z "$OCI_REGION" ]; then
        local account_id=$(aws sts get-caller-identity --query Account --output text)
        local ecr_secret="repo-${account_id}-${AWS_REGION}"
        if ! kubectl -n $namespace get secret $ecr_secret >/dev/null 2>&1; then
            echo "Applying temporary ECR credential secret $ecr_secret in namespace $namespace."
            local ecr_token=$(aws ecr get-login-password --region $AWS_REGION) || { echo "Failed to get ECR token"; exit 24; }
            echo "apiVersion: v1
kind: Secret
metadata:
  name: ${ecr_secret}
  namespace: ${namespace}
  labels:
    argocd.argoproj.io/secret-type: repo-creds
stringData:
  type: helm
  enableOCI: \"true\"
  url: ${account_id}.dkr.ecr.${AWS_REGION}.amazonaws.com
  username: AWS
  password: \"${ecr_token}\"" | kubectl apply -f - || { echo "Failed to create ECR credential secret $ecr_secret"; exit 24; }
        fi
    fi
    # Register credential-less OCI registries found in application files
    # ArgoCD requires a repository entry with enableOCI even for public OCI registries
    local done_urls=""
    for app_file in ./*.yaml; do
        for url in $(yq -r '.spec.sources[]? | select(.chart != null) | .repoURL' $app_file); do
            url="${url#oci://}"
            # Skip if already applied in this run
            case " $done_urls " in *" $url "*) continue;; esac
            # Skip if covered by a credential from the env variables
            local has_creds="false"
            for var in "${!GIT_AUTH_SOURCE_@}"; do
                local source="${!var}"
                source="${source#oci://}"
                if [[ "$url" == "$source"* ]]; then
                    has_creds="true"
                    break
                fi
            done
            if [ "$has_creds" == "true" ]; then
                continue
            fi
            # Sanitize url into a valid secret name
            local secret_name="repo-oci-$(echo $url | tr 'A-Z' 'a-z' | tr -c 'a-z0-9\n' '-' | sed 's/-*$//')"
            echo "Applying public OCI repository secret $secret_name in namespace $namespace."
            echo "apiVersion: v1
kind: Secret
metadata:
  name: ${secret_name}
  namespace: ${namespace}
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: helm
  enableOCI: \"true\"
  name: ${secret_name}
  url: ${url}" | kubectl apply -f - || { echo "Failed to create repository secret $secret_name"; exit 24; }
            done_urls="$done_urls $url"
        done
    done
}

# Login helm to an OCI registry when matching GIT_AUTH_* credentials exist
# Usage: helm_oci_login <repoURL without oci:// prefix>
helm_oci_login() {
    local repo="$1"
    for var in "${!GIT_AUTH_SOURCE_@}"; do
        local source="${!var}"
        source="${source#oci://}"
        if [[ "$repo" == "$source"* ]]; then
            local PASSWORD="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_PASSWORD/g')"
            local USERNAME="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_USERNAME/g')"
            # Registry host is the part before the first slash
            helm registry login "${repo%%/*}" --username "${!USERNAME}" --password "${!PASSWORD}" || { echo "Helm registry login failed for ${repo%%/*}"; exit 25; }
            return
        fi
    done
    # OCI_REGION must be checked before AWS_REGION: the OCI s3-compatible backend
    # also sets AWS_REGION. Oracle OCIR logins happen via the GIT_AUTH_SOURCE_* oci://
    # match above (helm registry login), so no credential helper config is needed.
    if [ -n "$OCI_REGION" ]; then
      :
    elif [ -n "$AWS_REGION" ]; then
      # Get current account number
      ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
      mkdir -p "$HOME/.config/helm/registry"
      cat > "$HOME/.config/helm/registry/config.json" <<EOF
{
  "credHelpers": {
    "${ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com": "ecr-login"
  }
}
EOF
    elif [ ! -z "$GOOGLE_REGION" ]; then
      mkdir -p "$HOME/.config/helm/registry"
      cat > "$HOME/.config/helm/registry/config.json" <<EOF
{
  "credHelpers": {
    "${GOOGLE_REGION}-docker.pkg.dev": "gcloud"
  }
}
EOF
    fi
}

# Bootstrap ArgoCD using Helm when not yet installed, supports GIT and OCI sources
# Returns "true" if bootstrap was performed, "false" otherwise
argocd_helm_bootstrap() {
    if [ "$ARGOCD_HOSTNAME" != "" ]; then
        echo "false"
        return
    fi

    echo "Detecting ArgoCD modules." >&2
    local did_bootstrap="false"

    for app_file in ./*.yaml; do
        # Detect source type of the ArgoCD module
        local source_type=""
        if yq -r '.spec.sources[0].path' $app_file | grep -q "modules/k8s/argocd"; then
            source_type="git"
        elif [ "$(yq -r '.spec.sources[0].chart' $app_file)" == "argocd" ]; then
            source_type="oci"
        else
            continue
        fi

        echo "Found $app_file ($source_type source), installing using helm." >&2
        local app=$(yq -r '.metadata.name' $app_file)
        # Ensure temporary files are removed on any failure exit
        trap "rm -rf values-$app.yaml git-$app oci-$app" EXIT
        yq -r '.spec.sources[0].helm.values' $app_file > values-$app.yaml
        local namespace=$(yq -r '.spec.destination.namespace' $app_file)
        local version=$(yq -r '.spec.sources[0].targetRevision' $app_file)
        local repo=$(yq -r '.spec.sources[0].repoURL' $app_file)
        local chart_dir=""

        if [ "$source_type" == "git" ]; then
            git_login
            local path=$(yq -r '.spec.sources[0].path' $app_file)
            git clone --depth 1 --single-branch --branch $version $repo git-$app >&2 || { echo "Git clone failed for $repo"; exit 25; }
            chart_dir="git-$app/$path"
        else
            # Pull and unpack the OCI chart so the value files inside the package can be referenced
            local chart=$(yq -r '.spec.sources[0].chart' $app_file)
            local oci_repo="${repo#oci://}"
            local chart_ref="oci://${oci_repo}/${chart}"
            local version_arg="--version $version"
            # Digest pinned revisions go into the chart reference, --version only accepts semver
            if [[ "$version" == sha256:* ]]; then
                chart_ref="${chart_ref}@${version}"
                version_arg=""
            fi
            helm_oci_login "$oci_repo" >&2
            helm pull $chart_ref $version_arg --untar --untardir oci-$app >&2 || { echo "Helm pull failed for $chart_ref"; exit 25; }
            chart_dir="oci-$app/$chart"
        fi

        # Apply namespace manifest with labels
        echo "apiVersion: v1
kind: Namespace
metadata:
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/warn: restricted
    tenancy.entigo.com/zone: infralib
  name: $namespace" | kubectl apply -f - || { echo "Failed to create namespace $namespace"; exit 22; }

        # Create repository secrets before install so ArgoCD can access the repositories
        argocd_repositories "$namespace" >&2

        helm upgrade --install -n $namespace \
            -f $chart_dir/values.yaml \
            -f $chart_dir/values-${PROVIDER}.yaml \
            -f values-$app.yaml \
            --set-string 'argocd.configs.cm.admin\.enabled=true' \
            --set argocd.server.ingress.enabled=false \
            --set argocd.server.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-server \
            --set argocd.dex.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-dex-server \
            --set argocd.redis.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-redis \
            --set argocd.repoServer.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-repo-server \
            --set argocd.applicationSet.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-applicationset-controller \
            --set argocd.notifications.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-notifications-controller \
            --set argocd.controller.statefulsetAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/StatefulSet:$app/$app-application-controller \
            --set argocd-apps.enabled=false $app $chart_dir >&2 || { echo "Helm install failed for $app in namespace $namespace"; exit 23; }
        rm -rf values-$app.yaml git-$app oci-$app >&2
        did_bootstrap="true"
    done

    # Refresh hostname after bootstrap
    export ARGOCD_HOSTNAME=$(get_argocd_hostname)

    echo "$did_bootstrap"
}

# Run argocd-plan for all yaml files in parallel
argocd_plan() {
    local helm_bootstrap=$(argocd_helm_bootstrap | tail -1)

    if [ "$ARGOCD_HOSTNAME" == "" ]; then
        export USE_ARGOCD_CLI="false"
        echo "Unable to get ArgoCD hostname. Falling back to kubectl."
    else
        # Sync repository secrets on every run so credential changes take effect
        argocd_repositories
    fi

    rm -f *.sync *.log
    PIDS=""
    for app_file in ./*.yaml; do
        argocd-apps-plan.sh $app_file > $app_file.log 2>&1 &
        PIDS="$PIDS $!=$app_file"
    done

    wait_for_jobs
    print_job_logs

    local ADD=$(cat ./*.log | grep "^Status " | grep -ve"Status:Synced" | grep -ve "Missing:0" | wc -l)
    local CHANGE=$(cat ./*.log | grep "^Status " | grep -ve"Status:Synced" | grep -ve "Changed:0" | wc -l)
    local DESTROY=$(cat ./*.log | grep "^Status " | grep -ve"Status:Synced" | grep -ve "RequiresPruning:0" | wc -l)
    # Prevent agent from confirming first bootstrap when ArgoCD's own application will always show changes
    if [ "$helm_bootstrap" == "true" -a $CHANGE -gt 0 ]; then
        CHANGE=$((CHANGE - 1))
    fi
    echo "ArgoCD Applications: ${ADD} to add, ${CHANGE} to change, ${DESTROY} to destroy."
    rm -f *.log

    if [ ! -z "$FAIL" ]; then
        echo "Failed jobs were:"
        for p in $FAIL; do
            echo "  - $(basename $(echo $p | cut -d"=" -f2))"
        done
        echo "Plan ArgoCD failed!"
        exit 21
    fi
}

# Run argocd-apply for all yaml files, respecting priority order
argocd_apply() {
    if [ "$ARGOCD_HOSTNAME" == "" ]; then
        export USE_ARGOCD_CLI="false"
        echo "Unable to get ArgoCD hostname. Falling back to kubectl."
    fi

    # Show priority summary
    echo "Application sync order:"
    declare -a file_priority_pairs
    declare -a unique_priorities
    for app_file in ./*.yaml; do
        priority=$(get_priority_from_yaml "$app_file")
        file_priority_pairs+=("$priority:$(basename $app_file)")
        if [[ ! " ${unique_priorities[@]} " =~ " ${priority} " ]]; then
            unique_priorities+=("$priority")
        fi
    done

    # Sort by priority and display
    printf '%s\n' "${file_priority_pairs[@]}" | sort -n | while IFS=':' read -r priority filename; do
        echo "  $filename: priority $priority"
    done
    echo ""

    for priority in $(printf '%s\n' "${unique_priorities[@]}" | sort -n); do
        echo "Syncing apps with priority $priority"
        PIDS=""
        for app_file in ./*.yaml; do
            app_priority=$(get_priority_from_yaml "$app_file")
            if [ $priority -eq $app_priority ]; then
                argocd-apps-apply.sh $app_file > $app_file.log 2>&1 &
                PIDS="$PIDS $!=$app_file"
            fi
        done

        wait_for_jobs

        # Retry failed apps
        if [ ! -z "$FAIL" ]; then
            echo "Retry Failed jobs:"
            for p in $FAIL; do
                echo "  - $(basename $(echo $p | cut -d"=" -f2))"
            done

            PIDS=""
            for p in $FAIL; do
                name=$(echo $p | cut -d"=" -f2)
                argocd-apps-apply.sh $name > $name.log 2>&1 &
                PIDS="$PIDS $!=$name"
            done

            wait_for_jobs
        fi

        handle_failed_jobs "Apply ArgoCD failed!" 21
    done
    rm -f *.log
}

# Plan destroy for all ArgoCD applications
argocd_plan_destroy() {
    rm -f *.sync *.log
    PIDS=""
    for app_file in ./*.yaml; do
        argocd-apps-plan-destroy.sh $app_file > $app_file.log 2>&1 &
        PIDS="$PIDS $!=$app_file"
    done

    wait_for_jobs

    local DESTROY=0
    for p in $COMPLETED; do
        name=$(echo $p | cut -d"=" -f2)
        cat $name.log
        let DESTROY++
    done

    for p in $FAIL; do
        name=$(echo $p | cut -d"=" -f2)
        cat $name.log
    done

    echo "ArgoCD Applications: 0 to add, 0 to change, ${DESTROY} to destroy."
    rm -f *.log

    if [ ! -z "$FAIL" ]; then
        echo "Failed jobs were:"
        for p in $FAIL; do
            echo "  - $(basename $(echo $p | cut -d"=" -f2))"
        done
        echo "Plan ArgoCD destroy failed!"
        exit 21
    fi
}

# Apply destroy for all ArgoCD applications
argocd_apply_destroy() {
    # Patch external-dns to sync mode before destroy
    echo "Detecting external-dns modules."
    for app_file in ./*.yaml; do
        # Module name comes from .path for git sources and from .chart for OCI sources
        if yq -r '.spec.sources[0].path // .spec.sources[0].chart // ""' $app_file | grep -Eq "(^|/)external-dns$"; then
            app=$(yq -r '.metadata.name' $app_file)
            echo "Found $app, patching policy to sync."
            POLICY_INDEX=$(kubectl get deployment $app -n $app -o jsonpath='{.spec.template.spec.containers[0].args}' | jq -r 'to_entries[] | select(.value | test("--policy")) | .key')
            if [ "$POLICY_INDEX" = "null" ] || [ -z "$POLICY_INDEX" ]; then
                kubectl patch deployment $app -n $app --type='json' -p='[
                    {
                        "op": "add",
                        "path": "/spec/template/spec/containers/0/args/-",
                        "value": "--policy=sync"
                    }
                ]'
            else
                kubectl patch deployment $app -n $app --type='json' -p='[
                    {
                        "op": "replace",
                        "path": "/spec/template/spec/containers/0/args/'$POLICY_INDEX'",
                        "value": "--policy=sync"
                    }
                ]'
            fi
        fi
    done

    # Show priority summary
    echo "Application sync order:"
    declare -a file_priority_pairs
    declare -a unique_priorities
    for app_file in ./*.yaml; do
        priority=$(get_priority_from_yaml "$app_file")
        file_priority_pairs+=("$priority:$(basename $app_file)")
        if [[ ! " ${unique_priorities[@]} " =~ " ${priority} " ]]; then
            unique_priorities+=("$priority")
        fi
    done

    # Sort by priority and display (reverse order for destroy)
    printf '%s\n' "${file_priority_pairs[@]}" | sort -nr | while IFS=':' read -r priority filename; do
        echo "  $filename: priority $priority"
    done
    echo ""

    for priority in $(printf '%s\n' "${unique_priorities[@]}" | sort -nr); do
        echo "Deleting apps with priority $priority"
        PIDS=""
        for app_file in ./*.yaml; do
            app_priority=$(get_priority_from_yaml "$app_file")
            if [ $priority -eq $app_priority ]; then
                argocd-apps-apply-destroy.sh $app_file > $app_file.log 2>&1 &
                PIDS="$PIDS $!=$app_file"
            fi
        done

        wait_for_jobs
        handle_failed_jobs "Destroy ArgoCD failed!" 21
    done
    rm -f *.log
}
