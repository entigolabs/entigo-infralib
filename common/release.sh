#!/bin/bash
echo "Assemble configuration for main"
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
source generate_config.sh

prepare_agent

echo "sources:
    - url: https://github.com/entigolabs/entigo-infralib
      version: main
      force_version: true
steps:" > agents/config.yaml

if [ "$AWS_ACCESS_KEY_ID" != "" ]
then
  default_aws_conf ../modules
fi
if [ "$GOOGLE_CREDENTIALS" != "" ]
then
  default_google_conf ../modules
fi

run_agents

PIDS=""

if [ "$AWS_ACCESS_KEY_ID" != "" ]
then
  ../modules/aws/kms/test.sh &
  PIDS="$PIDS $!=kms"
  ../modules/aws/cost-alert/test.sh &
  PIDS="$PIDS $!=cost-alert"
  ../modules/aws/hello-world/test.sh &
  PIDS="$PIDS $!=hello-world"
  ../modules/aws/vpc/test.sh &
  PIDS="$PIDS $!=vpc"
  ../modules/aws/route53/test.sh &
  PIDS="$PIDS $!=route53"
  ../modules/aws/eks/test.sh &
  PIDS="$PIDS $!=eks"
  ../modules/aws/crossplane/test.sh &
  PIDS="$PIDS $!=crossplane"
  ../modules/aws/ec2/test.sh &
  PIDS="$PIDS $!=ec2"
fi

FAIL=0
for p in $PIDS; do
    pid=$(echo $p | cut -d"=" -f1)
    name=$(echo $p | cut -d"=" -f2)
    wait $pid || let "FAIL+=1"
    echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]
then
  echo "FAILED GOLANG TEST $FAIL"
  exit 1
fi


