#!/bin/bash
if [ "$TESTING_VERSION" == "" ]
then
  TESTING_VERSION="latest" 
fi

MODULE_PATH="$(pwd)"
MODULETYPE=$(basename $(dirname $(pwd)))
MODULENAME=$(basename $(pwd))

if [ "$AWS_REGION" == "" ]
then
  echo "Defaulting AWS_REGION to eu-north-1"
  export AWS_REGION="eu-north-1"
fi

if [ "$PR_BRANCH" != "" ]
then
STEP_NAME="`whoami`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`-$MODULENAME"
else
STEP_NAME="`whoami`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`-$MODULENAME"
fi

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..

if [ "$1" == "testonly" ]
then
  for test in $(ls -1 $MODULE_PATH/test/*.yaml)
  do 
        testname=`basename $test | sed 's/\.yaml$//'`
        if [[ $STEP_NAME == *-rd-419-* ]]  #Change to runner-main- later
        then
          STEP_NAME=$(cat "agents/${MODULETYPE}_${testname}/config.yaml" | yq -r ".steps[] | select(.modules[].source == \"$MODULETYPE/$MODULENAME\") | .name")
          break
        fi
  done
else
  source common/generate_config.sh
  prepare_agent
  echo "sources:
 - url: /conf
steps:" > agents/config.yaml
  if [ "$AWS_ACCESS_KEY_ID" != "" ]
  then
    default_aws_conf
  fi
  if [ "$GOOGLE_CREDENTIALS" != "" ]
  then
    default_google_conf
  fi
  
  
  PIDS=""
  for test in $(ls -1 $MODULE_PATH/test/*.yaml)
  do 
        testname=`basename $test | sed 's/\.yaml$//'`
        
        if [[ $STEP_NAME == runner-rd-419-* ]]  #Change to runner-main- later
        then
          STEP_NAME=$(cat "agents/${MODULETYPE}_${testname}/config.yaml" | yq -r ".steps[] | select(.modules[].source == \"$MODULETYPE/$MODULENAME\") | .name")
        fi
        if ! yq '.steps[].name' "agents/${MODULETYPE}_${testname}/config.yaml" | grep -q "$STEP_NAME"
        then
            if [ "$MODULENAME" == "vpc" -o "$MODULENAME" == "cost-alert" ]
            then
              yq -y -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "approve": "force", "modules": [{"name": "'"$MODULENAME"'", "source": "'"$MODULETYPE"'/'"$MODULENAME"'"}]}]' "agents/${MODULETYPE}_${testname}/config.yaml"
            else
              yq -y -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "approve": "force", "vpc": {"attach": true}, "modules": [{"name": "'"$MODULENAME"'", "source": "'"$MODULETYPE"'/'"$MODULENAME"'"}]}]' "agents/${MODULETYPE}_${testname}/config.yaml"
            fi
        fi
        mkdir -p "agents/${MODULETYPE}_${testname}/config/$STEP_NAME"
        cp "$test" "agents/${MODULETYPE}_${testname}/config/$STEP_NAME/$MODULENAME.yaml"
        docker run --rm -v "$(pwd)":"/conf" -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e AWS_SESSION_TOKEN -w /conf --entrypoint ei-agent entigolabs/entigo-infralib-testing run -c /conf/agents/${MODULETYPE}_${testname}/config.yaml --steps "$STEP_NAME" --pipeline-type=local --prefix $testname &
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
        $TIMEOUT_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app entigolabs/entigo-infralib-testing:$TESTING_VERSION
 
