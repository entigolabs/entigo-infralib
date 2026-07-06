#!/bin/bash
#set -x

# Allow interactive shell access
if [ "$1" == "sh" ] || [ "$1" == "bash" ]; then
    exec "$@"
fi

[ -z $COMMAND ] && echo "COMMAND must be set" && exit 1

# Determine cloud provider and source appropriate functions
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
source "$SCRIPT_DIR/functions-common.sh"

if [ ! -z "$GOOGLE_REGION" ]; then
    source "$SCRIPT_DIR/functions-google.sh"
elif [ ! -z "$AWS_REGION" ]; then
    source "$SCRIPT_DIR/functions-aws.sh"
else
    echo "AWS_REGION or GOOGLE_REGION must be set"
    exit 1
fi

source "$SCRIPT_DIR/functions-argocd.sh"

# Validate required variables for non-local mode
if [ "$LOCAL_MODE" != "true" ]; then
    [ -z $TF_VAR_prefix ] && echo "TF_VAR_prefix must be set" && exit 1
    [ -z $INFRALIB_BUCKET ] && echo "INFRALIB_BUCKET must be set" && exit 1
fi

export TF_IN_AUTOMATION=1

# Execute command
case "$COMMAND" in
    test)
        run_tests
        ;;
    plan)
        prepare_plan_stage
        prepare_terraform
        terraform_plan
        upload_plan_artifact
        ;;
    plan-destroy)
        prepare_plan_stage
        prepare_terraform
        terraform_plan_destroy
        upload_plan_artifact
        ;;
    apply)
        fetch_plan_artifact
        prepare_terraform
        terraform_apply
        ;;
    apply-destroy)
        fetch_plan_artifact
        prepare_terraform
        terraform_apply_destroy
        ;;
    argocd-plan)
        prepare_plan_stage
        init_argocd_connection
        argocd_plan
        upload_plan_artifact
        ;;
    argocd-plan-destroy)
        prepare_plan_stage
        init_argocd_connection
        argocd_plan_destroy
        upload_plan_artifact
        ;;
    argocd-apply)
        fetch_plan_artifact
        init_argocd_connection
        argocd_apply
        ;;
    argocd-apply-destroy)
        fetch_plan_artifact
        init_argocd_connection
        argocd_apply_destroy
        ;;
    *)
        echo "Unknown command: $COMMAND"
        exit 1
        ;;
esac
