#!/bin/bash

export ENTIGO_INFRALIB_IMAGE="entigolabs/entigo-infralib-testing:v1.14.14-rc104"
export TFLINT_IMAGE="ghcr.io/terraform-linters/tflint:v0.50.3"
export KUBESCORE_IMAGE="martivo/kube-score:latest"

prepare_agent() {
  if [ -d agents ]
  then
    rm -rf agents
  fi
  mkdir agents
}

google_auth_login() {
  if [ "$CLOUDSDK_CONFIG" == "" ]
  then
    echo "Defaulting CLOUDSDK_CONFIG to $(echo ~)/.config/gcloud"
    export CLOUDSDK_CONFIG="$(echo ~)/.config/gcloud"
  fi
  if [ "$GOOGLE_CREDENTIALS" != "" -a ! -f $CLOUDSDK_CONFIG/application_default_credentials.json ]
  then
    echo "Found GOOGLE_CREDENTIALS, creating $CLOUDSDK_CONFIG/application_default_credentials.json"
    mkdir -p $CLOUDSDK_CONFIG
    echo ${GOOGLE_CREDENTIALS} > $CLOUDSDK_CONFIG/application_default_credentials.json
    gaccount=""
    attempt=1
    while [ -z "$gaccount" ] && [ "$attempt" -le "7" ]; do
      gcloud auth activate-service-account --key-file=$CLOUDSDK_CONFIG/application_default_credentials.json
      gcloud config set project $GOOGLE_PROJECT
      gcloud auth list
      gaccount=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")
      echo "Value for gaccount is '$gaccount'"
      if [ -z "$gaccount" ]
      then
        sleep 1.$((RANDOM % 9))
        echo "WARNING $attempt: Failed to retrieve expected result for: gcloud auth list --filter=status:ACTIVE"
        attempt=$((attempt + 1))
      fi
    done
    gcloud config set account $gaccount
  fi

}

#main means it runs in github and is applying the main branch
get_branch_name() {
  if [ "$PR_BRANCH" != "" ]
  then
    BRANCH=`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2 | cut -c1-7`
  else
    BRANCH=`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2 | cut -c1-7`
  fi
  if [ "`whoami`" == "runner" -a "$BRANCH" == "main" ]
  then
    BRANCH="main"
  else
    BRANCH="`whoami | cut -c1-4`-$BRANCH"
  fi
}

get_step_name_tf() {
  STEP_NAME="${BRANCH}-${MODULE_NAME}"
  if [ "$MODULE_NAME" == "config-rules" ] || [ "$MODULE_NAME" == "tgw-attach" ] || [ "$MODULE_NAME" == "dns" ]; then
    STEP_NAME="net"
  fi
}

get_step_name_k8s() {
  if [ "$BRANCH" == "main" ]
  then
    STEP_NAME="apps"
  else
    STEP_NAME=$APP_NAME
  fi
}

get_app_name() {
        if [ "$MODULE_NAME" == "crossplane-core" ]
        then
          APP_NAME="crossplane-system"
        elif [ "$MODULE_NAME" == "istio-istiod" ]
        then
          APP_NAME="istio-system"  
        elif [ "$MODULE_NAME" == "crossplane-aws" -o "$MODULE_NAME" == "crossplane-k8s" -o "$MODULE_NAME" == "crossplane-google" -o "$MODULE_NAME" == "google-gateway" -o "$MODULE_NAME" == "platform-apis" -o "$MODULE_NAME" == "platform-sql" ]
        then
          APP_NAME=$MODULE_NAME
        elif [ "$MODULE_NAME" == "argocd" -o "$MODULE_NAME" == "aws-alb" -o "$MODULE_NAME" == "external-secrets" -o "$MODULE_NAME" == "external-dns" -o "$MODULE_NAME" == "istio-base" -o "$MODULE_NAME" == "istio-gateway" -o "$MODULE_NAME" == "prometheus" -o "$MODULE_NAME" == "aws-storageclass" -o "$MODULE_NAME" == "entigo-portal-agent" -o "$MODULE_NAME" == "entigo-vulnerability-agent" -o "$MODULE_NAME" == "karpenter" -o "$MODULE_NAME" == "saml-proxy" -o "$MODULE_NAME" == "trivy" -o "$MODULE_NAME" == "kyverno" ]
        then
          APP_NAME="${MODULE_NAME}-$prefix"
        elif [ "$BRANCH" == "main" ]
        then
          APP_NAME="${MODULE_NAME}-$prefix"
        else
          APP_NAME="$BRANCH-$MODULE_NAME-$prefix"
        fi
}

generate_config() {
    local cloud=$1
    local step=$2
    shift 2
    local modules=("$@")
    local existing_step=""
    for module in "${modules[@]}"
    do
      module_name=$(basename $module)
      for test in $(ls -1 ./modules/$module/test/*.yaml)
      do 
        testname=`basename $test | sed 's/\.yaml$//'`
        if [ ! -f agents/${cloud}_$testname/config.yaml ]
        then
          mkdir -p agents/${cloud}_$testname/config
          cp agents/config.yaml agents/${cloud}_$testname/config.yaml
          local firstloop=1
        fi
        if [[ "$existing_step" != *"${step}-${testname}"* ]];
        then
          local existing_step="$existing_step ${step}-${testname}"
          if [ "${step}" == "net" ]
          then
          echo "    - name: ${step}
      type: terraform
      manual_approve_update: never
      manual_approve_run: never
      modules:" >> agents/${cloud}_$testname/config.yaml
          else
          echo "    - name: ${step}
      type: terraform
      manual_approve_update: never
      manual_approve_run: never
      vpc:
        attach: true
      modules:" >> agents/${cloud}_$testname/config.yaml
          fi
        fi

          echo "      - name: $module_name
        source: $module" >> agents/${cloud}_$testname/config.yaml
        mkdir -p agents/${cloud}_$testname/config/${step}
        cp $test agents/${cloud}_$testname/config/${step}/${module_name}.yaml
      done
    done
}

generate_config_k8s() {
    local MODULE_PATHS=$1
    local step=$2
    shift 2
    local modules=("$@")
    local existing_step=""
    BRANCH="main"
    for test in $(find $MODULE_PATHS  -maxdepth 1 -mindepth 1 -type d -exec basename {} \; | sort)
    do 
      MODULE_NAME=`basename $test`
      
      if [ ${#modules[@]} -eq 0 ] || [[ " ${modules[*]} " =~ " $MODULE_NAME " ]]
      then
      
      for cloud in $(find agents  -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
      do
        prefix="$(echo ${cloud} | cut -d'_' -f2)"
        if [ -f "$MODULE_PATHS/$MODULE_NAME/test/${cloud}.yaml" ]
        then
          get_app_name
          if [[ "$existing_step" != *"${cloud}"* ]];
          then
            mkdir -p agents/${cloud}/config/apps
            local existing_step="$existing_step ${cloud}"
            echo "    - name: apps
      type: argocd-apps
      manual_approve_update: never
      manual_approve_run: never
      argocd_namespace: argocd-$prefix
      modules:" >> agents/${cloud}/config.yaml
          fi

          echo "      - name: $APP_NAME
        source: $MODULE_NAME" >> agents/${cloud}/config.yaml
          mkdir -p agents/${cloud}/config/apps
          cp "$MODULE_PATHS/$MODULE_NAME/test/${cloud}.yaml" "agents/${cloud}/config/apps/${APP_NAME}.yaml"
        fi
      done
      fi
    done
}


run_agents() {
  google_auth_login
  
  local only_steps="$1"
  AGENT_OPTS=""
  if [ "$only_steps" != "" ]
  then
    AGENT_OPTS="--steps $only_steps"
  fi
  PIDS=""
  for agent in $(find ./agents -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
  do
    if [[ $agent == google_* ]]
    then
      if [ "$GOOGLE_REGION" == "" ]
      then
        echo "Defaulting GOOGLE_REGION to europe-north1"
        export GOOGLE_REGION="europe-north1"
      fi

      if [ "$GOOGLE_ZONE" == "" ]
      then
        echo "Defaulting GOOGLE_ZONE to europe-north1-a"
        export GOOGLE_ZONE="europe-north1-a"
      fi

      if [ "$GOOGLE_PROJECT" == "" ]
      then
        echo "Defaulting GOOGLE_PROJECT to entigo-infralib2"
        export GOOGLE_PROJECT="entigo-infralib2"
      fi

      docker run --rm -v $CLOUDSDK_CONFIG:/root/.config/gcloud -v $CLOUDSDK_CONFIG:/home/runner/.config/gcloud -v "$(pwd)":"/conf" -e LOCATION="$GOOGLE_REGION" -e ZONE="$GOOGLE_ZONE" -e PROJECT_ID="$GOOGLE_PROJECT" -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/$agent/config.yaml --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local $AGENT_OPTS  & 
      PIDS="$PIDS $!=$agent"
    elif [[ $agent == aws_* ]]
    then
        if [ "$(echo $agent | cut -d"_" -f2)" == "us" ]
        then
          echo "Defaulting AWS_REGION to us-east-1"
          export AGENT_AWS_REGION="us-east-1"
        else
          echo "Defaulting AWS_REGION to eu-north-1"
          export AGENT_AWS_REGION="eu-north-1"
        fi
        if [ $agent != "aws_spoke" -o \( $agent == "aws_spoke" -a "$AGENT_OPTS" == "" \) ]
        then
          docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION=$AGENT_AWS_REGION -e AWS_SESSION_TOKEN -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/$agent/config.yaml --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local $AGENT_OPTS &
          PIDS="$PIDS $!=$agent"
        fi
    else
      echo "Unknown cloud provider type $agent"
    fi
  done

  FAIL=""
  for p in $PIDS; do
      pid=$(echo $p | cut -d"=" -f1)
      name=$(echo $p | cut -d"=" -f2)
      wait $pid || FAIL="$FAIL $p"
      if [[ $FAIL == *$p* ]]
      then
        echo "$p Failed"
      else
        echo "$p Done"
      fi
  done
  if [ "$FAIL" != "" ]
  then
    echo "FAILED AGENT RUNS $FAIL"
    exit 1
  fi

}


test_tf() {
  PIDS=""
  if [ "$AWS_REGION" != "" ]
  then
    export AWS_REGION="eu-north-1"
    ./modules/aws/kms/test.sh testonly &
    PIDS="$PIDS $!=kms"
    ./modules/aws/cost-alert/test.sh testonly &
    PIDS="$PIDS $!=cost-alert"
    ./modules/aws/hello-world/test.sh testonly &
    PIDS="$PIDS $!=hello-world"
    ./modules/aws/vpc/test.sh testonly &
    PIDS="$PIDS $!=vpc"
    ./modules/aws/tgw-attach/test.sh testonly &
    PIDS="$PIDS $!=tgw-attach"
    ./modules/aws-v2/route53/test.sh testonly &
    PIDS="$PIDS $!=route53"
    ./modules/aws/route53-resolver-associate/test.sh testonly &
    PIDS="$PIDS $!=route53-resolver-associate"
    ./modules/aws/eks/test.sh testonly &
    PIDS="$PIDS $!=eks"
    ./modules/aws/eks-node-group/test.sh testonly &
    PIDS="$PIDS $!=eks-node-group"
    ./modules/aws/crossplane/test.sh testonly &
    PIDS="$PIDS $!=crossplane"
    ./modules/aws/ec2/test.sh testonly &
    PIDS="$PIDS $!=ec2"
    ./modules/aws/karpenter-node-role/test.sh testonly &
    PIDS="$PIDS $!=karpenter-node-role"
    ./modules/aws/config-rules/test.sh testonly &
    PIDS="$PIDS $!=config-rules"
  fi
  if [ "$GOOGLE_REGION" != "" ]
  then
    ./modules/google/services/test.sh testonly &
    PIDS="$PIDS $!=services"
    ./modules/google/vpc/test.sh testonly &
    PIDS="$PIDS $!=vpc"
    ./modules/google/dns/test.sh testonly &
    PIDS="$PIDS $!=dns"
    ./modules/google/gke/test.sh testonly &
    PIDS="$PIDS $!=gke"
    ./modules/google/crossplane/test.sh testonly &
    PIDS="$PIDS $!=crossplane"
  fi
  
  FAIL=""
  for p in $PIDS; do
      pid=$(echo $p | cut -d"=" -f1)
      name=$(echo $p | cut -d"=" -f2)
      wait $pid || FAIL="$FAIL $p"
      if [[ $FAIL == *$p* ]]
      then
        echo "$p Failed"
      else
        echo "$p Done"
      fi
  done
  if [ "$FAIL" != "" ]
  then
    echo "FAILED TF GOLANG TESTS $FAIL"
    exit 1
  fi

}

test_k8s() {
  google_auth_login
  
  gcloud container clusters get-credentials pri-infra-gke --region $GOOGLE_REGION
  gcloud container clusters get-credentials biz-infra-gke --region $GOOGLE_REGION
  aws eks update-kubeconfig --region $AWS_REGION --name pri-infra-eks
  aws eks update-kubeconfig --region $AWS_REGION --name biz-infra-eks
  
  PIDS=""
  #common
  ./modules/k8s/hello-world/test.sh testonly &
  PIDS="$PIDS $!=hello-world"
  ./modules/k8s/crossplane-core/test.sh testonly &
  PIDS="$PIDS $!=crossplane-core"
  
  #aws specific
  ./modules/k8s/crossplane-aws/test.sh testonly &
  PIDS="$PIDS $!=crossplane-aws"
  ./modules/k8s/aws-alb/test.sh testonly &
  PIDS="$PIDS $!=aws-alb"
  ./modules/k8s/aws-storageclass/test.sh testonly &
  PIDS="$PIDS $!=aws-storageclass"
  ./modules/k8s/cluster-autoscaler/test.sh testonly &
  PIDS="$PIDS $!=cluster-autoscaler"
  ./modules/k8s/entigo-portal-agent/test.sh testonly &
  PIDS="$PIDS $!=entigo-portal-agent"
  ./modules/k8s/entigo-vulnerability-agent/test.sh testonly &
  PIDS="$PIDS $!=entigo-vulnerability-agent"
  ./modules/k8s/metrics-server/test.sh testonly &
  PIDS="$PIDS $!=metrics-server"
  #google specific
  ./modules/k8s/crossplane-google/test.sh testonly &
  PIDS="$PIDS $!=crossplane-google"
  ./modules/k8s/google-gateway/test.sh testonly &
  PIDS="$PIDS $!=google-gateway"
  #common
  ./modules/k8s/crossplane-k8s/test.sh testonly &
  PIDS="$PIDS $!=crossplane-k8s"
  ./modules/k8s/crossplane-sql/test.sh testonly &
  PIDS="$PIDS $!=crossplane-sql"
  ./modules/k8s/external-dns/test.sh testonly &
  PIDS="$PIDS $!=external-dns"
  ./modules/k8s/external-secrets/test.sh testonly &
  PIDS="$PIDS $!=external-secrets"
  ./modules/k8s/argocd/test.sh testonly &
  PIDS="$PIDS $!=argocd"
  ./modules/k8s/istio-base/test.sh testonly &
  PIDS="$PIDS $!=istio-base"
  ./modules/k8s/istio-gateway/test.sh testonly &
  PIDS="$PIDS $!=istio-gateway"
  ./modules/k8s/istio-istiod/test.sh testonly &
  PIDS="$PIDS $!=istio-istiod"
  ./modules/k8s/karpenter/test.sh testonly &
  PIDS="$PIDS $!=karpenter"
  ./modules/k8s/kiali/test.sh testonly &
  PIDS="$PIDS $!=kiali"
  ./modules/k8s/loki/test.sh testonly &
  PIDS="$PIDS $!=loki"
  ./modules/k8s/mimir/test.sh testonly &
  PIDS="$PIDS $!=mimir"
  ./modules/k8s/prometheus/test.sh testonly &
  PIDS="$PIDS $!=prometheus"
  ./modules/k8s/promtail/test.sh testonly &
  PIDS="$PIDS $!=promtail"
  ./modules/k8s/grafana/test.sh testonly &
  PIDS="$PIDS $!=grafana"
  ./modules/k8s/harbor/test.sh testonly &
  PIDS="$PIDS $!=harbor"
  ./modules/k8s/trivy/test.sh testonly &
  PIDS="$PIDS $!=trivy"
  
  FAIL=""
  for p in $PIDS; do
      pid=$(echo $p | cut -d"=" -f1)
      name=$(echo $p | cut -d"=" -f2)
      wait $pid || FAIL="$FAIL $p"
      if [[ $FAIL == *$p* ]]
      then
        echo "$p Failed"
      else
        echo "$p Done"
      fi
  done
  if [ "$FAIL" != "" ]
  then
    echo "FAILED K8S GOLANG TESTS $FAIL"
    exit 2
  fi
}


default_aws_conf() {
  generate_config "aws" "net" "aws/kms" "aws/cost-alert" "aws/hello-world" "aws/vpc" "aws/tgw-attach" "aws-v2/route53" "aws/route53-resolver-associate" "aws/ecr-proxy" "aws/config-rules"
  generate_config "aws" "infra" "aws/eks" "aws/eks-node-group" "aws/crossplane" "aws/ec2" "aws/karpenter-node-role"
}

default_google_conf() {
  generate_config "google" "net" "google/services" "google/vpc" "google/dns" "google/gar-proxy"
  generate_config "google" "infra" "google/gke" "google/crossplane"
}

full_k8s_conf() {
  generate_config_k8s "./modules/k8s" "apps"
}

main_k8s_conf() {
  generate_config_k8s "./modules/k8s" "apps" "argocd" "aws-alb" "aws-storageclass" "cluster-autoscaler" "crossplane-aws" "crossplane-core" "crossplane-google" "crossplane-sql" "external-dns" "external-secrets" "google-gateway" "istio-base" "istio-istiod" "loki" "metrics-server" "rbac-bindings"
}
