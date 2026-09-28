#!/bin/bash
#set -x
if [ "$COMMAND" == "test" ]
then
    exec entrypoint-core.sh "$@"
elif [ -n "$AWS_REGION" ] || [ -n "$GOOGLE_REGION" ] || [ -n "$OCI_REGION" ] || [ -n "$AZURE_SUBSCRIPTION_ID" ]; then
    exec /usr/bin/ei-agent provision "$@"
else
    echo "AWS_REGION, GOOGLE_REGION, OCI_REGION or AZURE_SUBSCRIPTION_ID must be set"
    exit 1
fi
