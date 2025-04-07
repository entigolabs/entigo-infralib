#!/bin/bash

if [ "$AWS_REGION" == "" ]
then
  echo "Defaulting AWS_REGION to eu-north-1"
  export AWS_REGION="eu-north-1"
fi

MODULE_PATH="$(pwd)"
MODULE_TYPE=$(basename $(dirname $(pwd)))
MODULE_NAME=$(basename $(pwd))

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..
source common/generate_config.sh

get_branch_name
get_step_name_tf

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
  echo "callback:
    url: http://localhost
    key: 123456
sources:
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
           yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "manual_approve_update": "never", "manual_approve_run": "never", "modules": [{"name": "'"$MODULE_NAME"'", "source": "'"$MODULE_TYPE"'/'"$MODULE_NAME"'"}]}]' "agents/${MODULE_TYPE}_${testname}/config.yaml"
        fi
        mkdir -p "agents/${MODULE_TYPE}_${testname}/config/$STEP_NAME"
        cp "$test" "agents/${MODULE_TYPE}_${testname}/config/$STEP_NAME/$MODULE_NAME.yaml"
        docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/${MODULE_TYPE}_${testname}/config.yaml --steps "$STEP_NAME" --pipeline-type=local --prefix $testname &
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

docker run -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
	-e AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN" \
	-e AWS_REGION="$AWS_REGION" \
	-e COMMAND="test" \
	-e STEP_NAME="$STEP_NAME" \
        $TIMEOUT_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app $ENTIGO_INFRALIB_IMAGE
 
