#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/.."
cd $SCRIPTPATH || exit 1

DOCKER_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
DOCKER_OPTS="-it"
fi
echo "Doing docker run $DOCKER_OPTS --rm -v"
docker run $DOCKER_OPTS --rm -v "$(pwd)":"/data" ghcr.io/terraform-linters/tflint-bundle:v0.46.1.1 
