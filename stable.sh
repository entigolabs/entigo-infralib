#!/bin/bash
echo "Assemble configuration for stable"
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
source common/generate_config.sh

prepare_agent

echo "sources:
    - url: https://github.com/entigolabs/entigo-infralib-release
      version: stable
steps:" > agents/config.yaml

if [ "$AWS_ACCESS_KEY_ID" != "" ]
then
  default_aws_conf
fi
if [ "$GOOGLE_CREDENTIALS" != "" ]
then
  default_google_conf
fi

run_agents

test_all
