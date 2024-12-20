#!/bin/bash
if [ "$TESTING_VERSION" == "" ]
then
  TESTING_VERSION="v1.2.8-rc27"
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

if [ "$ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS" == "" ]
then
  echo "Defaulting ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS to true"
  export ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS="true"
fi

if [ "$ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS" == "" ]
then
  echo "Defaulting ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS to true"
  export ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS="true"
fi

DOCKER_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
  DOCKER_OPTS="-it"
  export CLOUDSDK_CONFIG="$(echo ~)/.config/gcloud"
else
  #This is needed for terratest terraform execution
  DOCKER_OPTS='-e GOOGLE_CREDENTIALS'
  export CLOUDSDK_CONFIG="$(echo ~)/.config/gcloud-$(tr -dc A-Za-z0-9 </dev/urandom | head -c 6; echo)"
  #This is needed for terratest bucket creation
  mkdir -p $CLOUDSDK_CONFIG
  gaccount=""
  attempt=1
  while [ -z "$gaccount" ] && [ "$attempt" -le "7" ]; do
    echo ${GOOGLE_CREDENTIALS} > $CLOUDSDK_CONFIG/application_default_credentials.json
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
	-e ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS="$ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS" \
	-e ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS="$ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS" \
  -v $CLOUDSDK_CONFIG:/root/.config/gcloud \
	$TIMEOUT_OPTS $DOCKER_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app entigolabs/entigo-infralib-testing:$TESTING_VERSION
 
