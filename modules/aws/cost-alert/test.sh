#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
AWS_REGION="us-east-1"
exec $SCRIPTPATH/../../../common/test.sh "$@"
