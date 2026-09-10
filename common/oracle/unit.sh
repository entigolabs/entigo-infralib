#!/bin/bash

if [ "$OCI_REGION" == "" ]
then
  echo "Defaulting OCI_REGION to eu-frankfurt-1"
  export OCI_REGION="eu-frankfurt-1"
fi

if [ "$ORACLE_COMPARTMENT_ID" == "" ]
then
  echo "ERROR: ORACLE_COMPARTMENT_ID should be set to the compartment used for testing."
  exit 5
fi

if [ "$OCI_CONFIG_FILE" == "" ]
then
  echo "Defaulting OCI_CONFIG_FILE to $(echo ~)/.oci/config"
  export OCI_CONFIG_FILE="$(echo ~)/.oci/config"
fi

MODULE_PATH="$(pwd)"
MODULE_TYPE_VERSIONED=$(basename $(dirname $(pwd)))
MODULE_TYPE=$(echo $(basename $(dirname $(pwd))) | cut -d"-" -f1)
MODULE_NAME=$(basename $(pwd))

SCRIPTPATH=$(dirname "$0")
cd $SCRIPTPATH/../..
source common/generate_config.sh

get_branch_name
get_step_name_tf_oracle

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
  default_oracle_conf

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
        docker run --rm -v "$OCI_CONFIG_FILE":"$OCI_CONFIG_FILE":ro -v "$(pwd)":"/conf" -e OCI_CONFIG_FILE="$OCI_CONFIG_FILE" -e OCI_REGION="$OCI_REGION" -e ORACLE_COMPARTMENT_ID="$ORACLE_COMPARTMENT_ID" -e ORACLE_REGION="$OCI_REGION" -w /conf --entrypoint ei-agent $ENTIGO_INFRALIB_IMAGE run -c /conf/agents/${MODULE_TYPE}_${testname}/config.yaml --steps "$STEP_NAME" --pipeline-type=local --prefix $testname &
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

DOCKER_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
  DOCKER_OPTS="-it"
fi

docker run -e OCI_REGION="$OCI_REGION" \
	-e ORACLE_COMPARTMENT_ID="$ORACLE_COMPARTMENT_ID" \
	-e OCI_CONFIG_FILE="$OCI_CONFIG_FILE" \
	-e COMMAND="test" \
	-e STEP_NAME="$STEP_NAME" \
	-v "$OCI_CONFIG_FILE":"$OCI_CONFIG_FILE":ro \
        $TIMEOUT_OPTS $DOCKER_OPTS --rm -v "$(pwd)":"/app" -v "$(pwd)/../../../common":"/common" -w /app $ENTIGO_INFRALIB_IMAGE
