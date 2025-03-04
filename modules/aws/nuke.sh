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
delete_versions_in_batch() {
    local bucket_name="$1"
    
    # Batch delete object versions
    aws s3api list-object-versions --bucket "$bucket_name" --output json | \
    jq -c '.Versions[], .DeleteMarkers[] | select(.Key != null) | {Key: .Key, VersionId: .VersionId}' | \
    split -l 1000 - /tmp/versions_batch_
    
    # Process each batch file
    for batch_file in /tmp/versions_batch_*; do
        # Create a JSON file for batch delete
        jq -n --arg bucket "$bucket_name" \
            '{Bucket: $bucket, Delete: {Objects: inputs}}' \
            "$batch_file" > "${batch_file}_delete.json"
        
        # Perform batch delete (quietly)
        aws s3api batch-delete-objects \
            --bucket "$bucket_name" \
            --delete "file://${batch_file}_delete.json" \
            > /dev/null 2>&1
        
        # Clean up temporary files
        rm "$batch_file" "${batch_file}_delete.json"
    done
}

# List and process all buckets
echo "Listing and preparing to delete all versions from S3 buckets:"
aws s3 ls | while read -r line; do
    # Extract bucket name (3rd column in the ls output)
    bucket=$(echo "$line" | awk '{print $3}')
    
    # Confirm before processing each bucket
    echo "Delete ALL versions from bucket $bucket"
    delete_versions_in_batch "$bucket"
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

