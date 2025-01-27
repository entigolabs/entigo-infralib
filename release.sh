#!/bin/bash
echo "Assemble configuration for main"
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
source common/generate_config.sh

prepare_agent

echo "sources:
    - url: https://github.com/entigolabs/entigo-infralib
      version: main
      force_version: true
steps:" > agents/config.yaml

if [ "$AWS_REGION" != "" ]
then
  default_aws_conf
fi
if [ "$GOOGLE_REGION" != "" ]
then
  default_google_conf
fi

default_k8s_conf

run_agents

test_tf
