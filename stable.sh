#!/bin/bash
echo "Assemble configuration for stable"
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
source common/generate_config.sh

prepare_agent

echo "sources:
    - url: https://github.com/entigolabs/entigo-infralib-release
steps:" > agents/config.yaml

if [ "$AWS_REGION" != "" ]
then
  default_aws_conf
fi
if [ "$GOOGLE_REGION" != "" ]
then
  default_google_conf
fi

if [ "$1" == "" ]
then
  full_k8s_conf
  run_agents
  test_tf
  test_k8s

elif [ "$1" == "tf" ]
then
  main_k8s_conf
  docker pull $ENTIGO_INFRALIB_IMAGE
  run_agents
  docker pull $TFLINT_IMAGE
  test_tf

elif  [ "$1" == "k8s" ]
then
  full_k8s_conf
  docker pull $ENTIGO_INFRALIB_IMAGE
  rm -rf agents/aws_ext agents/aws_min agents/aws_us
  run_agents apps
  docker pull $KUBESCORE_IMAGE
  test_k8s
fi
