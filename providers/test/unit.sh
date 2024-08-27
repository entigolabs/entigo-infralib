#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/.."
cd $SCRIPTPATH || exit 1

if [ "$PR_BRANCH" != "" ]
then
prefix="`whoami`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`"
else
prefix="`whoami`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`"
fi

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

DOCKER_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
  DOCKER_OPTS="-it"
else
  #This is needed for terratest terraform execution
  DOCKER_OPTS='-e GOOGLE_CREDENTIALS'
  #This is needed for terratest bucket creation
  mkdir -p $(echo ~)/.config/gcloud 
  echo ${GOOGLE_CREDENTIALS} > $(echo ~)/.config/gcloud/application_default_credentials.json
fi

docker run -e GOOGLE_REGION="$GOOGLE_REGION" \
	-e GOOGLE_ZONE="$GOOGLE_ZONE" \
	-e GOOGLE_PROJECT="$GOOGLE_PROJECT" \
	-e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
	-e AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN" \
	-e AWS_REGION="$AWS_REGION" \
	-e TF_VAR_prefix="$prefix" \
	-e ENTIGO_INFRALIB_DESTROY="$ENTIGO_INFRALIB_DESTROY" \
	-e ENTIGO_INFRALIB_TEST_TIMEOUT="60m" \
	-v $(echo ~)/.config/gcloud/application_default_credentials.json:/root/.config/gcloud/application_default_credentials.json \
       	 $DOCKER_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../common":"/common" -w /app entigolabs/entigo-infralib-testing:v0.13.12-rc14
