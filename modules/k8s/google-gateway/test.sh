#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
export ENTIGO_INFRALIB_KUBECTL_AWS_CONTEXTS="false"
exec $SCRIPTPATH/../../../common/test.sh "$@"
