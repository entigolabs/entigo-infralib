#!/bin/bash

prepare_agent() {
  if [ -d agents ]
  then
    rm -rf agents
  fi
  mkdir agents
  
  if [ "$AWS_REGION" == "" ]
  then
    echo "Defaulting AWS_REGION to eu-north-1"
    export AWS_REGION="eu-north-1"
  fi
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
  cd agents
  for agent in $(find . -maxdepth 1 -mindepth 1 -type d -printf "%f\n")
  do
    cd $agent
    if [[ $agent == google_* ]]
    then
      echo "skip google for now"
      #docker run -it --rm -v "$(pwd)":"/conf" -e PROJECT_ID -e LOCATION -e ZONE --entrypoint ei-agent entigolabs/entigo-infralib-testing run -c /conf/config.yaml --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local  
    elif [[ $agent == aws_* ]]
    then
      if [ "$module" != "" ]
      then
        step=$(cat config.yaml| yq -r ".steps[] | select(.modules[].name == \"$module\") | .name")
        if [ "$step" != "" ]
        then
          echo "Module $module found in step $step"
          docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN --entrypoint ei-agent entigolabs/entigo-infralib-testing run -c /conf/config.yaml --steps $step --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local &
          PIDS="$PIDS $!=$agent"
        fi
      else
        docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN --entrypoint ei-agent entigolabs/entigo-infralib-testing run -c /conf/config.yaml --prefix $(echo $agent | cut -d"_" -f2) --pipeline-type=local &
        PIDS="$PIDS $!=$agent"
      fi
      
    else
      echo "Unknown cloud provider type $agent"
    fi
    cd ..
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
  cd ..

}


default_aws_conf() {
  generate_config "$1/aws" "net" "kms" "cost-alert" "hello-world" "vpc" "route53"
  generate_config "$1/aws" "infra" "eks" "crossplane" "ec2"
}

default_google_conf() {
  generate_config "$1/google" "net" "services" "vpc" 
  generate_config "$1/google" "infra" "gke" "dns" "crossplane"
}
