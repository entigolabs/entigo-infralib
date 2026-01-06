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
        TES_CONNECTION=$(argocd --server ${ARGOCD_HOSTNAME} --grpc-web app list)
        if [ $? -eq 0 ]; then
            echo "Connected to ArgoCD successfully."
            export USE_ARGOCD_CLI="true"
        fi
    fi
}

# Bootstrap ArgoCD using Helm when not yet installed
# Returns "true" if bootstrap was performed, "false" otherwise
argocd_helm_bootstrap() {
    if [ "$ARGOCD_HOSTNAME" != "" ]; then
        echo "false"
        return
    fi

    git_login
    echo "Detecting ArgoCD modules." >&2
    local did_bootstrap="false"

    for app_file in ./*.yaml; do
        if yq -r '.spec.sources[0].path' $app_file | grep -q "modules/k8s/argocd"; then
            echo "Found $app_file, installing using helm." >&2
            local app=$(yq -r '.metadata.name' $app_file)
            yq -r '.spec.sources[0].helm.values' $app_file > values-$app.yaml
            local namespace=$(yq -r '.spec.destination.namespace' $app_file)
            local version=$(yq -r '.spec.sources[0].targetRevision' $app_file)
            local repo=$(yq -r '.spec.sources[0].repoURL' $app_file)
            local path=$(yq -r '.spec.sources[0].path' $app_file)
            git clone --depth 1 --single-branch --branch $version $repo git-$app >&2

            # Create bootstrap value file that is only used first time ArgoCD is created
            if compgen -A variable | grep -q "^GIT_AUTH_SOURCE_"; then
                echo "
argocd:
  configs:
    repositories:" > git-$app/$path/extra_repos.yaml

                for var in "${!GIT_AUTH_SOURCE_@}"; do
                    NAME=$(echo $var | sed 's/GIT_AUTH_SOURCE_//g')
                    SOURCE="$(echo ${!var})"
                    PASSWORD="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_PASSWORD/g')"
                    USERNAME="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_USERNAME/g')"
                    echo "      ${NAME}:
        url: ${SOURCE}.git
        name: ${NAME}
        password: ${!PASSWORD}
        username: ${!USERNAME}" >> git-$app/$path/extra_repos.yaml
                done
            else
                touch git-$app/$path/extra_repos.yaml
            fi

            helm upgrade --create-namespace --install -n $namespace \
                -f git-$app/$path/values.yaml \
                -f git-$app/$path/values-${PROVIDER}.yaml \
                -f values-$app.yaml \
                -f git-$app/$path/extra_repos.yaml \
                --set-string 'argocd.configs.cm.admin\.enabled=true' \
                --set argocd.server.ingress.enabled=false \
                --set argocd.server.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-server \
                --set argocd.dex.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-dex-server \
                --set argocd.redis.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-redis \
                --set argocd.repoServer.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-repo-server \
                --set argocd.applicationSet.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-applicationset-controller \
                --set argocd.notifications.deploymentAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/Deployment:$app/$app-notifications-controller \
                --set argocd.controller.statefulsetAnnotations."argocd\.argoproj\.io/tracking-id"=$app:apps/StatefulSet:$app/$app-application-controller \
                --set argocd-apps.enabled=false $app git-$app/$path >&2
            rm -rf values-$app.yaml git-$app >&2
            did_bootstrap="true"
        fi
    done

    # Refresh hostname after bootstrap
    export ARGOCD_HOSTNAME=$(get_argocd_hostname)

    echo "$did_bootstrap"
}

# Run argocd-plan for all yaml files in parallel
argocd_plan() {
    local helm_bootstrap=$(argocd_helm_bootstrap)

    if [ "$ARGOCD_HOSTNAME" == "" ]; then
        export USE_ARGOCD_CLI="false"
        echo "Unable to get ArgoCD hostname. Falling back to kubectl."
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
        if yq -r '.spec.sources[0].path' $app_file | grep -q "modules/k8s/external-dns"; then
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
