#!/bin/bash

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


run_agents() {
  local module="$1"
  PIDS=""
  for agent in $(find ./agents -maxdepth 1 -mindepth 1 -type d -printf "%f\n")
  do
    if [[ $agent == google_* ]]
    then
      if [ "$LOCATION" == "" ]
      then
        echo "Defaulting GOOGLE_REGION to europe-north1"
        export LOCATION="europe-north1"
      fi

      if [ "$ZONE" == "" ]
      then
        echo "Defaulting GOOGLE_ZONE to europe-north1-a"
        export ZONE="europe-north1-a"
      fi

      if [ "$PROJECT_ID" == "" ]
      then
        echo "Defaulting PROJECT_ID to entigo-infralib2"
        export PROJECT_ID="entigo-infralib2"
      fi
      if [ "$CLOUDSDK_CONFIG" == "" ]
      then
        export CLOUDSDK_CONFIG="$(echo ~)/.config/gcloud"
      fi
      
      if [ "$GOOGLE_CREDENTIALS" != "" ]
      then
        mkdir -p $CLOUDSDK_CONFIG
        echo ${GOOGLE_CREDENTIALS} > $CLOUDSDK_CONFIG/application_default_credentials.json
      fi
      docker run --rm -v $CLOUDSDK_CONFIG:/root/.config/gcloud -v "$(pwd)":"/conf" -e PROJECT_ID -e LOCATION -e ZONE -w /conf --entrypoint ei-agent entigolabs/entigo-infralib-testing:agent-alpha1 run -c /conf/agents/$agent/config.yaml --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local &
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
    
        docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN -w /conf --entrypoint ei-agent entigolabs/entigo-infralib-testing:agent-alpha1 run -c /conf/agents/$agent/config.yaml --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local &
        PIDS="$PIDS $!=$agent"
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
