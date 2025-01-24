#!/bin/bash

export ENTIGO_INFRALIB_IMAGE="entigolabs/entigo-infralib-testing:a124"

prepare_agent() {
  if [ -d agents ]
  then
    rm -rf agents
  fi
  mkdir agents
}

generate_config() {
    local cloud=$(basename $1)
    local modulepath=$1
    local step=$2
    shift 2
    local modules=("$@")
    local existing_step=""
    for module in "${modules[@]}"
    do
      for test in $(ls -1 ${modulepath}/$module/test/*.yaml)
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
      approve: force
      modules:" >> agents/${cloud}_$testname/config.yaml
          else
          echo "    - name: ${step}
      type: terraform
      approve: force
      vpc:
        attach: true
      modules:" >> agents/${cloud}_$testname/config.yaml
          fi
        fi

          echo "      - name: $module
        source: ${cloud}/$module" >> agents/${cloud}_$testname/config.yaml
        mkdir -p agents/${cloud}_$testname/config/${step}
        cp $test agents/${cloud}_$testname/config/${step}/${module}.yaml
      done
    done
}

generate_config_k8s() {
    local modulepath=$1
    local step=$2
    local existing_step=""
    for test in $(find $modulepath  -maxdepth 1 -mindepth 1 -type d -printf "%f\n" | sort)
    do 
      module=`basename $test`
      for cloud in $(find agents  -maxdepth 1 -mindepth 1 -type d -printf "%f\n")
      do
        if [ -f "$modulepath/$module/test/${cloud}.yaml" ]
        then
          if [ "$module" == "crossplane-core" ]
          then
            module_name="crossplane-system"
          elif [ "$module" == "crossplane-aws" -o "$module" == "crossplane-k8s" -o "$module" == "crossplane-google" ]
          then
            module_name=$module
          elif [ "$module" == "istio-istiod" ]
          then
            module_name="istio-system"
            
          else
            module_name="$module-$(echo ${cloud} | cut -d'_' -f2)"
          fi
          if [[ "$existing_step" != *"${cloud}"* ]];
          then
            mkdir -p agents/${cloud}/config/apps
            local existing_step="$existing_step ${cloud}"
            echo "    - name: apps
      type: argocd-apps
      approve: force
      argocd_namespace: $module_name
      kubernetes_cluster_name: '{{ .toutput.eks.cluster_name }}'
      modules:" >> agents/${cloud}/config.yaml
          fi

          echo "      - name: $module_name
        source: $module" >> agents/${cloud}/config.yaml
          mkdir -p agents/${cloud}/config/apps
          cp "$modulepath/$module/test/${cloud}.yaml" "agents/${cloud}/config/apps/${module_name}.yaml"
        fi
      done
    done
}


run_agents() {
  #docker pull $ENTIGO_INFRALIB_IMAGE
  if [ "$CLOUDSDK_CONFIG" == "" ]
  then
    echo "Defaulting CLOUDSDK_CONFIG to $(echo ~)/.config/gcloud"
    export CLOUDSDK_CONFIG="$(echo ~)/.config/gcloud"
  fi
  if [ "$GOOGLE_CREDENTIALS" != "" ]
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
  
  local module="$1"
  PIDS=""
  for agent in $(find ./agents -maxdepth 1 -mindepth 1 -type d -printf "%f\n")
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

      docker run --rm -v $CLOUDSDK_CONFIG:/root/.config/gcloud -v $CLOUDSDK_CONFIG:/home/runner/.config/gcloud -v "$(pwd)":"/conf" -e LOCATION="$GOOGLE_REGION" -e ZONE="$GOOGLE_ZONE" -e PROJECT_ID="$GOOGLE_PROJECT" -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/$agent/config.yaml --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local &
      PIDS="$PIDS $!=$agent"
    elif [[ $agent == aws_* ]]
    then
        if [ "$(echo $agent | cut -d"_" -f2)" == "us" ]
        then
          echo "Defaulting AWS_REGION to us-east-1"
          export AWS_REGION="us-east-1"
        else
          echo "Defaulting AWS_REGION to eu-north-1"
          export AWS_REGION="eu-north-1"
        fi
    
        docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/$agent/config.yaml --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local --steps apps
        #PIDS="$PIDS $!=$agent"
    else
      echo "Unknown cloud provider type $agent"
    fi
  done

  FAIL=0
  for p in $PIDS; do
      pid=$(echo $p | cut -d"=" -f1)
      name=$(echo $p | cut -d"=" -f2)
      wait $pid || let "FAIL+=1"
      echo $p $FAIL
  done
  if [ "$FAIL" -ne 0 ]
  then
    echo "FAILED AGENT RUN $FAIL"
    exit 1
  fi

}


test_all() {
  PIDS=""
  if [ "$AWS_REGION" != "" ]
  then
    ./modules/aws/kms/test.sh testonly &
    PIDS="$PIDS $!=kms"
    ./modules/aws/cost-alert/test.sh testonly &
    PIDS="$PIDS $!=cost-alert"
    ./modules/aws/hello-world/test.sh testonly &
    PIDS="$PIDS $!=hello-world"
    ./modules/aws/vpc/test.sh testonly &
    PIDS="$PIDS $!=vpc"
    ./modules/aws/route53/test.sh testonly &
    PIDS="$PIDS $!=route53"
    ./modules/aws/eks/test.sh testonly &
    PIDS="$PIDS $!=eks"
    ./modules/aws/crossplane/test.sh testonly &
    PIDS="$PIDS $!=crossplane"
    ./modules/aws/ec2/test.sh testonly &
    PIDS="$PIDS $!=ec2"
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

  FAIL=0
  for p in $PIDS; do
      pid=$(echo $p | cut -d"=" -f1)
      name=$(echo $p | cut -d"=" -f2)
      wait $pid || let "FAIL+=1"
      echo $p $FAIL
  done
  if [ "$FAIL" -ne 0 ]
  then
    echo "FAILED GOLANG TEST $FAIL"
    exit 1
  fi


}


default_aws_conf() {
  generate_config "./modules/aws" "net" "kms" "cost-alert" "hello-world" "vpc" "route53"
  generate_config "./modules/aws" "infra" "eks" "crossplane" "ec2"
}

default_google_conf() {
  generate_config "./modules/google" "net" "services" "vpc" "dns"
  generate_config "./modules/google" "infra" "gke" "crossplane"
}


default_k8s_conf() {
  generate_config_k8s "./modules/k8s" "apps"
}
