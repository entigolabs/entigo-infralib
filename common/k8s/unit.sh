#!/bin/bash
if [ "$ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS" == "" ]
then
  export ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS="true"
fi

if [ "$ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS" == "" ]
then
  export ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS="true"
fi



MODULE_PATH="$(pwd)"


SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..
source common/generate_config.sh

DOCKER_OPTS=""
if [ "$GOOGLE_CREDENTIALS" != "" ]
then
    DOCKER_OPTS='-e GOOGLE_CREDENTIALS'
fi

google_auth_login

if [ "$1" != "testonly" ]
then
  
  if [ "`whoami`" == "runner" ]
  then
    if [ "$GOOGLE_CREDENTIALS" != "" ]
    then
      gcloud container clusters get-credentials pri-infra-gke --region $GOOGLE_REGION
      if [ $? -ne 0 ]
      then
        echo "Failed to get context for Google pri-infra-gke"
        exit 1
      fi
      gcloud container clusters get-credentials biz-infra-gke --region $GOOGLE_REGION
      if [ $? -ne 0 ]
      then
        echo "Failed to get context for Google biz-infra-gke"
        exit 1
      fi
    fi
    if [ "$AWS_ACCESS_KEY_ID" != "" ]
    then
      aws eks update-kubeconfig --region $AWS_REGION --name pri-infra-eks
      if [ $? -ne 0 ]
      then
        echo "Failed to get context for AWS pri-infra-gke"
        exit 1
      fi
      aws eks update-kubeconfig --region $AWS_REGION --name biz-infra-eks
      if [ $? -ne 0 ]
      then
        echo "Failed to get context for AWS biz-infra-gke"
        exit 1
      fi
    fi
    docker pull $ENTIGO_INFRALIB_IMAGE
  fi
  MODULE_NAME=$(basename $MODULE_PATH)
  get_branch_name
  prepare_agent
if [ "$BRANCH" == "main" ]
then
echo "callback:
    url: http://localhost
    key: 123456
sources:
  - url: https://github.com/entigolabs/entigo-infralib
      version: main
      force_version: true
steps:" > agents/config.yaml

else
  echo "callback:
    url: http://localhost
    key: 123456
sources:
  - url: /conf
steps:" > agents/config.yaml
fi
  if [ "$AWS_ACCESS_KEY_ID" != "" ]
  then
    default_aws_conf
  fi
  if [ -d "$CLOUDSDK_CONFIG" ]
  then
    default_google_conf
  fi
  full_k8s_conf
  
  MODULE_NAME=$(basename $MODULE_PATH)
  get_branch_name
  PIDS=""
  for test in $(ls -1 $MODULE_PATH/test/*.yaml | grep -ve"static_tests.yaml")
  do 
        testname=`basename $test | sed 's/\.yaml$//'`
        prefix="$(echo $testname | cut -d"_" -f2)"
        get_app_name
        get_step_name_k8s
        if ! yq '.steps[].name' "agents/${testname}/config.yaml" | grep -q "$STEP_NAME"
        then
          yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "argocd-apps", "argocd_namespace":"argocd-'"$prefix"'", "manual_approve_update": "never", "manual_approve_run": "never", "modules": [{"name": "'"$APP_NAME"'", "source": "'"$MODULE_NAME"'"}]}]' "agents/${testname}/config.yaml"
        fi
        mkdir -p "agents/${testname}/config/$STEP_NAME"
        cp "$test" "agents/${testname}/config/$STEP_NAME/$APP_NAME.yaml"
        
      
        if [[ $testname == google_* ]]
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

          docker run --rm -v $CLOUDSDK_CONFIG:/root/.config/gcloud -v $CLOUDSDK_CONFIG:/home/runner/.config/gcloud -v "$(pwd)":"/conf" -e LOCATION="$GOOGLE_REGION" -e ZONE="$GOOGLE_ZONE" -e PROJECT_ID="$GOOGLE_PROJECT" -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/$testname/config.yaml --prefix $prefix --pipeline-type=local --steps "$STEP_NAME" &
          PIDS="$PIDS $!=$testname"
        elif [[ $testname == aws_spoke ]]
        then
	   echo "Skip aws_spoke test"
        elif [[ $testname == aws_* ]]
        then
            if [ "$prefix" == "us" ]
            then
              echo "Defaulting AWS_REGION to us-east-1"
              export AWS_REGION="us-east-1"
            else
              echo "Defaulting AWS_REGION to eu-north-1"
              export AWS_REGION="eu-north-1"
            fi
            docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/$testname/config.yaml --prefix $prefix --pipeline-type=local --steps "$STEP_NAME" &
            PIDS="$PIDS $!=$testname"
        else
          echo "Unknown cloud provider type $testname"
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
fi

cd $MODULE_PATH
TIMEOUT_OPTS=""
if [ "$ENTIGO_INFRALIB_TEST_TIMEOUT" != "" ]
then
  TIMEOUT_OPTS="-e ENTIGO_INFRALIB_TEST_TIMEOUT=$ENTIGO_INFRALIB_TEST_TIMEOUT"
fi
prefix=""
MODULE_NAME=$(basename $MODULE_PATH)
get_branch_name
get_app_name

docker run -e GOOGLE_REGION="$GOOGLE_REGION" \
	-e GOOGLE_ZONE="$GOOGLE_ZONE" \
	-e GOOGLE_PROJECT="$GOOGLE_PROJECT" \
	-e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
	-e AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN" \
	-e AWS_REGION="$AWS_REGION" \
	-e COMMAND="test" \
	-e APP_NAME="$APP_NAME" \
  -v $CLOUDSDK_CONFIG:/root/.config/gcloud \
	$TIMEOUT_OPTS $DOCKER_OPTS --rm -v "$(echo ~)/.kube":"/root/.kube" -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app $ENTIGO_INFRALIB_IMAGE
 
