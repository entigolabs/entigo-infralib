#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH || exit 1



if [ "$AWS_REGION" == "" ]
then
  echo "Defaulting AWS_REGION to eu-north-1"
  export AWS_REGION="eu-north-1"
fi

echo "$SCRIPTPATH/aws-nuke-config.yml"

# Function to completely delete all versions of objects in a versioned bucket
delete_all_versions() {
    local bucket_name="$1"
    echo "Fully deleting all versions from bucket: $bucket_name"
    
    # List and delete all object versions
    aws s3api list-object-versions --bucket "$bucket_name" --output json | \
    jq -r '.Versions[], .DeleteMarkers[] | select(.Key != null) | [.Key, .VersionId] | @tsv' | \
    while IFS=$'\t' read -r key version_id; do
        echo "Deleting version: $key (Version ID: $version_id)"
        aws s3api delete-object --bucket "$bucket_name" --key "$key" --version-id "$version_id"
    done
    
    # Additional cleanup for any remaining delete markers
    aws s3api list-object-versions --bucket "$bucket_name" --output json | \
    jq -r '.DeleteMarkers[] | select(.Key != null) | [.Key, .VersionId] | @tsv' | \
    while IFS=$'\t' read -r key version_id; do
        echo "Removing delete marker: $key (Version ID: $version_id)"
        aws s3api delete-object --bucket "$bucket_name" --key "$key" --version-id "$version_id"
    done
}

# List and process all buckets
echo "Listing and preparing to delete all versions from S3 buckets:"
aws s3 ls | while read -r line; do
    # Extract bucket name (3rd column in the ls output)
    bucket=$(echo "$line" | awk '{print $3}')
    
    # Confirm before processing each bucket
    echo "Delete ALL versions from bucket $bucket"
    #delete_all_versions "$bucket"
done

echo "Versioned bucket cleanup process completed."


if [ "$GITHUB_ACTION" == "" ]
then

docker run -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
	-e AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN" \
	-e AWS_REGION="$AWS_REGION" \
	--rm -it -v "$SCRIPTPATH/aws-nuke-config.yml":"/home/aws-nuke/config.yml" ghcr.io/ekristen/aws-nuke:v3.48.2 run --config /home/aws-nuke/config.yml --access-key-id ${AWS_ACCESS_KEY_ID} --secret-access-key ${AWS_SECRET_ACCESS_KEY} --session-token ${AWS_SESSION_TOKEN} --no-dry-run

else

docker run -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
	-e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
	-e AWS_REGION="$AWS_REGION" \
	--rm -v "$SCRIPTPATH/aws-nuke-config.yml":"/home/aws-nuke/config.yml" ghcr.io/ekristen/aws-nuke:v3.48.2 run --config /home/aws-nuke/config.yml --access-key-id ${AWS_ACCESS_KEY_ID} --secret-access-key ${AWS_SECRET_ACCESS_KEY} --no-dry-run --force --max-wait-retries 100

fi

