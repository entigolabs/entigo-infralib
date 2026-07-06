#!/bin/bash
#set -x
if [ "$COMMAND" == "test" ]
then
    exec entrypoint-core.sh "$@"
elif [ -n "$GOOGLE_REGION" ]; then
    exec entrypoint-core.sh "$@"
elif [ -n "$AWS_REGION" ]; then
    exec /usr/bin/ei-agent provision "$@"
else
    echo "AWS_REGION or GOOGLE_REGION must be set"
    exit 1
fi
