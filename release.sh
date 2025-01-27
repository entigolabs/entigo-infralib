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

#When we run release in local we will run goole, aws and k8s tests all in one process. No argument needs to be supplied.
#In GitHub "Agent Release" we run google and aws in separate processes (the tf argument is supplied).
if [ "$1" == "tf" -o "$1" == "" ]
then
docker pull $ENTIGO_INFRALIB_IMAGE
run_agents
docker pull $TFLINT_IMAGE
test_tf
fi
#In GitHub "Agent Release" we run k8s tests in separate processes (the k8s argument is supplied). This will test k8s modules in aws and goole.
if [ "$1" == "k8s" -o "$1" == ""  ]
then
docker pull $KUBESCORE_IMAGE
test_k8s
fi
