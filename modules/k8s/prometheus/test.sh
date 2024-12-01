#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
#Use standard testing.
export KUBESCORE_EXTRA_OPTS="--ignore-test pod-probes"
exec ../../../common/test.sh $SCRIPTPATH
