#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
export KUBESCORE_EXTRA_OPTS="--ignore-test pod-networkpolicy"
export ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS="false"
exec $SCRIPTPATH/../../../common/test.sh "$@"
