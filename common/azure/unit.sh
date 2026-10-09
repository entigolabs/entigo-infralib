#!/bin/bash
MODULE_PATH="$(pwd)"
MODULE_TYPE_VERSIONED=$(basename $(dirname $(pwd)))
MODULE_TYPE=$(echo $(basename $(dirname $(pwd))) | cut -d"-" -f1)
MODULE_NAME=$(basename $(pwd))

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..
source common/generate_config.sh

if [ "$AZURE_SUBSCRIPTION_ID" == "" ]
then
  echo "ERROR: AZURE_SUBSCRIPTION_ID should be set."
  exit 5
fi

get_branch_name
get_step_name_tf_azure
if [ "$AZURE_LOCATION" == "" ]
then
  echo "Defaulting AZURE_LOCATION to westeurope"
  export AZURE_LOCATION="westeurope"
fi

if [ "$1" == "testonly" ]
then
  for test in $(ls -1 $MODULE_PATH/test/*.yaml)
  do
        testname=`basename $test | sed 's/\.yaml$//'`
        if [ "$BRANCH" == "main" ]
        then
          STEP_NAME=$(cat "agents/${MODULE_TYPE}_${testname}/config.yaml" | yq -r ".steps[] | select(.modules[].source == \"$MODULE_TYPE_VERSIONED/$MODULE_NAME\") | .name")
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
enable_opentofu: true
steps:" > agents/config.yaml
  default_azure_conf

  PIDS=""
  for test in $(ls -1 $MODULE_PATH/test/*.yaml)
  do
        testname=`basename $test | sed 's/\.yaml$//'`

        if [ "$BRANCH" == "main" ]
        then
          STEP_NAME=$(cat "agents/${MODULE_TYPE}_${testname}/config.yaml" | yq -r ".steps[] | select(.modules[].source == \"$MODULE_TYPE_VERSIONED/$MODULE_NAME\") | .name")
        fi
        if ! yq '.steps[].name' "agents/${MODULE_TYPE}_${testname}/config.yaml" | grep -q "$STEP_NAME"
        then
            if [ "$MODULE_NAME" == "vpc" ]
            then
              yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "manual_approve_update": "never", "manual_approve_run": "never", "modules": [{"name": "'"$MODULE_NAME"'", "source": "'"$MODULE_TYPE_VERSIONED"'/'"$MODULE_NAME"'"}]}]' "agents/${MODULE_TYPE}_${testname}/config.yaml"
            else
              yq -i '.steps += [{"name": "'"$STEP_NAME"'", "type": "terraform", "manual_approve_update": "never", "manual_approve_run": "never", "vpc": {"attach": true}, "modules": [{"name": "'"$MODULE_NAME"'", "source": "'"$MODULE_TYPE_VERSIONED"'/'"$MODULE_NAME"'"}]}]' "agents/${MODULE_TYPE}_${testname}/config.yaml"
            fi
        fi
        mkdir -p "agents/${MODULE_TYPE}_${testname}/config/$STEP_NAME"
        cp "$test" "agents/${MODULE_TYPE}_${testname}/config/$STEP_NAME/$MODULE_NAME.yaml"
        docker run --rm -v "$(pwd)":"/conf" -e AZURE_SUBSCRIPTION_ID -e AZURE_LOCATION -e AZURE_TENANT_ID -e AZURE_CLIENT_ID -e AZURE_CLIENT_SECRET -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/${MODULE_TYPE}_${testname}/config.yaml --steps "$STEP_NAME" --pipeline-type=local --prefix $testname &
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

docker run -e COMMAND="test" \
	-e STEP_NAME="$STEP_NAME" \
	-e AZURE_SUBSCRIPTION_ID -e AZURE_LOCATION -e AZURE_TENANT_ID -e AZURE_CLIENT_ID -e AZURE_CLIENT_SECRET $TIMEOUT_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app $ENTIGO_INFRALIB_IMAGE
