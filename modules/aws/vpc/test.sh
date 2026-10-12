#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
#Waits for flow log files to be delivered to S3
export ENTIGO_INFRALIB_TEST_TIMEOUT="45m"
cd $SCRIPTPATH
exec $SCRIPTPATH/../../../common/test.sh "$@"
