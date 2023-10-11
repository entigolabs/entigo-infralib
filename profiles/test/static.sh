#!/bin/bash 

export DOCKER_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
  export DOCKER_OPTS="-it"
fi

for line in `ls -1 *.yaml`
do
  echo "Static test of $line"
  if [ ! -f "test/$line" ]
  then
    echo "No patch config for profile $line found"
    exit 1
  fi
  docker run $DOCKER_OPTS --rm -v "$(pwd)/$line":"/etc/ei-agent/base.yaml" -v "$(pwd)/test/$line":"/etc/ei-agent/patch.yaml" entigolabs/entigo-infralib-agent ei-agent merge --base-config=/etc/ei-agent/base.yaml --config=/etc/ei-agent/patch.yaml
  if [ $? -ne 0 ]
  then
    exit 2
  fi
done
exit 0
