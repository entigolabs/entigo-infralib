#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
export KUBESCORE_EXTRA_OPTS="--ignore-test container-ephemeral-storage-request-and-limit --ignore-test container-resources"
exec $SCRIPTPATH/../../../common/test.sh "$@"
