#!/bin/bash

if [ "$AWS_REGION" == "" ]
then
  echo "Defaulting AWS_REGION to eu-north-1"
  export AWS_REGION="eu-north-1"
fi

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
APP_NAME="`whoami`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`-$MODULENAME"
else
APP_NAME="`whoami`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`-$MODULENAME"
fi

if [[ $APP_NAME == *-rd-419-* ]]  #Change to runner-main- later
then
  APP_NAME="$MODULENAME-$testname"
fi

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..
source common/generate_config.sh


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
 
