#!/bin/bash
if [ "$TESTING_VERSION" == "" ]
then
  TESTING_VERSION="v0.13.7-rc14"
fi

if [ "$PR_BRANCH" != "" ]
then
prefix="`whoami`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`"
else
prefix="`whoami`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`"
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
  DOCKER_OPTS="-it "
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

echo "If authentication fails use command 'gcloud auth application-default login', if you used different project then run also 'gcloud config set project entigo-infralib2' beforehand."

docker run -e GOOGLE_REGION="$GOOGLE_REGION" \
	-e GOOGLE_ZONE="$GOOGLE_ZONE" \
	-e GOOGLE_PROJECT="$GOOGLE_PROJECT" \
	-e TF_VAR_prefix="$prefix" \
	-e ENTIGO_INFRALIB_DESTROY="$ENTIGO_INFRALIB_DESTROY" \
	-v $(echo ~)/.config/gcloud:/root/.config/gcloud \
        $TIMEOUT_OPTS $DOCKER_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -v "$(pwd)/../../../providers":"/providers" -w /app entigolabs/entigo-infralib-testing:$TESTING_VERSION
 
