#!/bin/bash
if [ "$TESTING_VERSION" == "" ]
then
  TESTING_VERSION="v0.11.21-rc8"
fi

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

TIMEOUT_OPTS=""
if [ "$ENTIGO_INFRALIB_TEST_TIMEOUT" != "" ]
then
  TIMEOUT_OPTS="-e ENTIGO_INFRALIB_TEST_TIMEOUT=$ENTIGO_INFRALIB_TEST_TIMEOUT"
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
	-e ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS="true" \
	-e ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS="true" \
	-v $(echo ~)/.config/gcloud/application_default_credentials.json:/root/.config/gcloud/application_default_credentials.json \
	$TIMEOUT_OPTS $DOCKER_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app entigolabs/entigo-infralib-testing:$TESTING_VERSION
 
