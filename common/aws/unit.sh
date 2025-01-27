#!/bin/bash

if [ "$AWS_REGION" == "" ]
then
  echo "Defaulting AWS_REGION to eu-north-1"
  export AWS_REGION="eu-north-1"
fi

MODULE_PATH="$(pwd)"
MODULETYPE=$(basename $(dirname $(pwd)))
MODULENAME=$(basename $(pwd))

if [ "$PR_BRANCH" != "" ]
then
STEP_NAME="`whoami | cut -c1-4`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2 | cut -c1-4`-$MODULENAME"
else
STEP_NAME="`whoami | cut -c1-4`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2 | cut -c1-4`-$MODULENAME"
fi

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..
source common/generate_config.sh



if [ "$1" == "testonly" ]
then
  for test in $(ls -1 $MODULE_PATH/test/*.yaml)
  do 
        testname=`basename $test | sed 's/\.yaml$//'`
        if [[ $STEP_NAME == *-main-* ]]  #Change to *-main-* later
        then
          STEP_NAME=$(cat "agents/${MODULETYPE}_${testname}/config.yaml" | yq -r ".steps[] | select(.modules[].source == \"$MODULETYPE/$MODULENAME\") | .name")
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
        
        if [[ $STEP_NAME == runn-main-* ]]
        then
          STEP_NAME=$(cat "agents/${MODULETYPE}_${testname}/config.yaml" | yq -r ".steps[] | select(.modules[].source == \"$MODULETYPE/$MODULENAME\") | .name")
        fi
        if ! yq '.steps[].name' "agents/${MODULETYPE}_${testname}/config.yaml" | grep -q "$STEP_NAME"
        then
            if [ "$MODULENAME" == "vpc" -o "$MODULENAME" == "cost-alert" ]
            then
              yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "approve": "force", "modules": [{"name": "'"$MODULENAME"'", "source": "'"$MODULETYPE"'/'"$MODULENAME"'"}]}]' "agents/${MODULETYPE}_${testname}/config.yaml"
            else
              yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "approve": "force", "vpc": {"attach": true}, "modules": [{"name": "'"$MODULENAME"'", "source": "'"$MODULETYPE"'/'"$MODULENAME"'"}]}]' "agents/${MODULETYPE}_${testname}/config.yaml"
            fi
        fi
        mkdir -p "agents/${MODULETYPE}_${testname}/config/$STEP_NAME"
        cp "$test" "agents/${MODULETYPE}_${testname}/config/$STEP_NAME/$MODULENAME.yaml"
        docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/${MODULETYPE}_${testname}/config.yaml --steps "$STEP_NAME" --pipeline-type=local --prefix $testname &
        PIDS="$PIDS $!=$testname"
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

docker run -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
	-e AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN" \
	-e AWS_REGION="$AWS_REGION" \
	-e COMMAND="test" \
	-e STEP_NAME="$STEP_NAME" \
        $TIMEOUT_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app $ENTIGO_INFRALIB_IMAGE
 
