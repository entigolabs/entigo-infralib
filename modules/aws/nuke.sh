#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH || exit 1



if [ "$AWS_REGION" == "" ]
then
  echo "Defaulting AWS_REGION to eu-north-1"
  export AWS_REGION="eu-north-1"
fi

echo "$SCRIPTPATH/aws-nuke-config.yml"


if [ "$GITHUB_ACTION" == "" ]
then

docker run -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
	-e AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN" \
	-e AWS_REGION="$AWS_REGION" \
	--rm -it -v "$SCRIPTPATH/aws-nuke-config.yml":"/home/aws-nuke/config.yml" ghcr.io/ekristen/aws-nuke:v3.23.0 run --config /home/aws-nuke/config.yml --access-key-id ${AWS_ACCESS_KEY_ID} --secret-access-key ${AWS_SECRET_ACCESS_KEY} --session-token ${AWS_SESSION_TOKEN} --no-dry-run

else

docker run -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
	-e AWS_REGION="$AWS_REGION" \
	--rm -v "$SCRIPTPATH/aws-nuke-config.yml":"/home/aws-nuke/config.yml" ghcr.io/ekristen/aws-nuke:v3.23.0 run --config /home/aws-nuke/config.yml --access-key-id ${AWS_ACCESS_KEY_ID} --secret-access-key ${AWS_SECRET_ACCESS_KEY} --no-dry-run --force --max-wait-retries 100

fi

