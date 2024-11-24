#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
#Use standard testing.
export KUBESCORE_EXTRA_OPTS="--ignore-test container-ephemeral-storage-request-and-limit --ignore-test container-resources"
exec ../../../common/test.sh $SCRIPTPATH
