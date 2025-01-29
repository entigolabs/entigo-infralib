#!/bin/bash
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

MODULE_PATH="$(pwd)"
MODULE_TYPE=$(basename $(dirname $(pwd)))
MODULE_NAME=$(basename $(pwd))

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..
source common/generate_config.sh

get_branch_name
get_step_name_tf

google_auth_login


DOCKER_OPTS=""
if [ "$GOOGLE_CREDENTIALS" != "" ]
then
    DOCKER_OPTS='-e GOOGLE_CREDENTIALS'
fi


if [ "$1" == "testonly" ]
then
  for test in $(ls -1 $MODULE_PATH/test/*.yaml)
  do 
        testname=`basename $test | sed 's/\.yaml$//'`
        if [ "$BRANCH" == "main" ] 
        then
          STEP_NAME=$(cat "agents/${MODULE_TYPE}_${testname}/config.yaml" | yq -r ".steps[] | select(.modules[].source == \"$MODULE_TYPE/$MODULE_NAME\") | .name")
          break
        fi
  done
else
  if [ "`whoami`" == "runner" ]
  then
    docker pull $ENTIGO_INFRALIB_IMAGE
  fi

  prepare_agent
  echo "sources:
 - url: /conf
steps:" > agents/config.yaml
  if [ "$AWS_ACCESS_KEY_ID" != "" ]
  then
    default_aws_conf
  fi
  if [ "$CLOUDSDK_CONFIG" != "" ]
  then
    default_google_conf
  fi
  if [ "$AWS_ACCESS_KEY_ID" == "" -a "$CLOUDSDK_CONFIG" == "" ]
  then
    echo "ERROR: AWS_ACCESS_KEY_ID or CLOUDSDK_CONFIG should be set."
    exit 5
  fi
  
  PIDS=""
  for test in $(ls -1 $MODULE_PATH/test/*.yaml)
  do 
        testname=`basename $test | sed 's/\.yaml$//'`
        
        if [ "$BRANCH" == "main" ] 
        then
          STEP_NAME=$(cat "agents/${MODULE_TYPE}_${testname}/config.yaml" | yq -r ".steps[] | select(.modules[].source == \"$MODULE_TYPE/$MODULE_NAME\") | .name")
        fi
        if ! yq '.steps[].name' "agents/${MODULE_TYPE}_${testname}/config.yaml" | grep -q "$STEP_NAME"
        then
            if [ "$MODULE_NAME" == "vpc" -o "$MODULE_NAME" == "cost-alert" ]
            then
              yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "approve": "force", "modules": [{"name": "'"$MODULE_NAME"'", "source": "'"$MODULE_TYPE"'/'"$MODULE_NAME"'"}]}]' "agents/${MODULE_TYPE}_${testname}/config.yaml"
            else
              yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "approve": "force", "vpc": {"attach": true}, "modules": [{"name": "'"$MODULE_NAME"'", "source": "'"$MODULE_TYPE"'/'"$MODULE_NAME"'"}]}]' "agents/${MODULE_TYPE}_${testname}/config.yaml"
            fi
        fi
        mkdir -p "agents/${MODULE_TYPE}_${testname}/config/$STEP_NAME"
        cp "$test" "agents/${MODULE_TYPE}_${testname}/config/$STEP_NAME/$MODULE_NAME.yaml"
        docker run --rm -v $CLOUDSDK_CONFIG:/root/.config/gcloud -v $CLOUDSDK_CONFIG:/home/runner/.config/gcloud -v "$(pwd)":"/conf" -e LOCATION="$GOOGLE_REGION" -e ZONE="$GOOGLE_ZONE" -e PROJECT_ID="$GOOGLE_PROJECT" -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/${MODULE_TYPE}_${testname}/config.yaml --steps "$STEP_NAME" --pipeline-type=local --prefix $testname &
        PIDS="$PIDS $!=$testname"
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

echo "If authentication fails use command 'gcloud auth application-default login', if you used different project then run also 'gcloud config set project entigo-infralib2' beforehand."

docker run -e GOOGLE_REGION="$GOOGLE_REGION" \
	-e GOOGLE_ZONE="$GOOGLE_ZONE" \
	-e GOOGLE_PROJECT="$GOOGLE_PROJECT" \
	-e COMMAND="test" \
	-e STEP_NAME="$STEP_NAME" \
	-v $CLOUDSDK_CONFIG:/root/.config/gcloud \
        $TIMEOUT_OPTS $DOCKER_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app $ENTIGO_INFRALIB_IMAGE
 
