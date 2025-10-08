#!/bin/bash

if [ "$1" == "" ]; then
    echo "First parameter is provider version to pull and push"
    exit 1
fi

VERSION=$1
PROVIDERS="provider-family-aws provider-aws-elasticache provider-aws-ram provider-aws-networkfirewall provider-aws-ec2 provider-aws-route53 provider-aws-ecr provider-aws-sqs provider-aws-s3 provider-aws-iam provider-aws-vpc provider-aws-secretsmanager provider-aws-rds provider-aws-kms provider-aws-eks"

for provider in $PROVIDERS; do
    SOURCE="xpkg.upbound.io/upbound/$provider:$VERSION"
    DEST="entigolabs/$provider:$VERSION"
    
    echo "Copying $SOURCE to $DEST"
    docker buildx imagetools create --tag $DEST $SOURCE
    
    if [ $? -ne 0 ]; then
        echo "Failed to copy $SOURCE to $DEST"
        exit 2
    fi
done

echo "All providers copied successfully"


