#!/bin/bash

if [ "$CLOUDSDK_CONFIG" == "" ]
then
  echo "Defaulting CLOUDSDK_CONFIG to $(echo ~)/.config/gcloud"
  export CLOUDSDK_CONFIG="$(echo ~)/.config/gcloud"
fi

if [ "$ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS" == "" ]
then
  export ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS="true"
fi

if [ "$ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS" == "" ]
then
  export ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS="true"
fi

DOCKER_OPTS=""
if [ "$GOOGLE_CREDENTIALS" != "" -a ! -f "$CLOUDSDK_CONFIG/application_default_credentials.json" ]
then
    echo "Found GOOGLE_CREDENTIALS, creating $CLOUDSDK_CONFIG/application_default_credentials.json"
    #This is needed for terratest terraform execution
    DOCKER_OPTS='-e GOOGLE_CREDENTIALS'
    #This is needed for terratest bucket creation
    mkdir -p $CLOUDSDK_CONFIG
    echo ${GOOGLE_CREDENTIALS} > $CLOUDSDK_CONFIG/application_default_credentials.json
fi

MODULE_PATH="$(pwd)"
MODULETYPE=$(basename $(dirname $(pwd)))
MODULENAME=$(basename $(pwd))

if [ "$PR_BRANCH" != "" ]
then
APP_NAME="`whoami | cut -c1-4`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2 | cut -c1-7`-$MODULENAME"
STEP_NAME="`whoami | cut -c1-4`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2 | cut -c1-7`"
else
APP_NAME="`whoami | cut -c1-4`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2 | cut -c1-7`-$MODULENAME"
STEP_NAME="`whoami | cut -c1-4`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2 | cut -c1-7`"
fi

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..
source common/generate_config.sh

if [ "$1" != "testonly" ]
then

  prepare_agent
  echo "sources:
 - url: /conf
steps:" > agents/config.yaml
  if [ "$AWS_ACCESS_KEY_ID" != "" ]
  then
    default_aws_conf
  fi
  if [ -d "$CLOUDSDK_CONFIG" ]
  then
    default_google_conf
  fi
  default_k8s_conf
  
  PIDS=""
  for test in $(ls -1 $MODULE_PATH/test/*.yaml)
  do 
        testname=`basename $test | sed 's/\.yaml$//'`
        if [ $STEP_NAME == "runn-main" -o "$MODULENAME" == "crossplane-core" -o "$MODULENAME" == "crossplane-aws" -o "$MODULENAME" == "crossplane-k8s" -o "$MODULENAME" == "crossplane-google" ]  #Change to *-main-* later
        then
          STEP_NAME="apps"
          APP_NAME=$MODULENAME
        fi
        if ! yq '.steps[].name' "agents/${testname}/config.yaml" | grep -q "$STEP_NAME"
        then
          yq --version
          yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "argocd-apps", "argocd_namespace":"argocd-'"$(echo $testname | cut -d"_" -f2)"'", "approve": "force", "modules": [{"name": "'"$APP_NAME"'", "source": "'"$MODULENAME"'"}]}]' "agents/${testname}/config.yaml"
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

          docker run --rm -v $CLOUDSDK_CONFIG:/root/.config/gcloud -v $CLOUDSDK_CONFIG:/home/runner/.config/gcloud -v "$(pwd)":"/conf" -e LOCATION="$GOOGLE_REGION" -e ZONE="$GOOGLE_ZONE" -e PROJECT_ID="$GOOGLE_PROJECT" -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/$testname/config.yaml --prefix $(echo $testname | cut -d"_" -f2) --pipeline-type=local --steps "$STEP_NAME" & #--steps apps
          PIDS="$PIDS $!=$testname"
        elif [[ $testname == aws_* ]]
        then
            if [ "$(echo $testname | cut -d"_" -f2)" == "us" ]
            then
              echo "Defaulting AWS_REGION to us-east-1"
              export AWS_REGION="us-east-1"
            else
              echo "Defaulting AWS_REGION to eu-north-1"
              export AWS_REGION="eu-north-1"
            fi
        
            docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/$testname/config.yaml --prefix $(echo $testname | cut -d"_" -f2) --pipeline-type=local --steps "$STEP_NAME" & #--steps apps
            PIDS="$PIDS $!=$testname"
        else
          echo "Unknown cloud provider type $testname"
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
fi

cd $MODULE_PATH
TIMEOUT_OPTS=""
if [ "$ENTIGO_INFRALIB_TEST_TIMEOUT" != "" ]
then
  TIMEOUT_OPTS="-e ENTIGO_INFRALIB_TEST_TIMEOUT=$ENTIGO_INFRALIB_TEST_TIMEOUT"
fi
pwd
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
 
