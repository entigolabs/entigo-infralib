#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
export KUBESCORE_EXTRA_OPTS="--ignore-test pod-networkpolicy --ignore-test container-ephemeral-storage-request-and-limit --ignore-test pod-probes"
exec $SCRIPTPATH/../../../common/test.sh "$@"
