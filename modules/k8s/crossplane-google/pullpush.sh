#!/bin/bash

if [ "$1" == "" ]; then
    echo "First parameter is provider version to pull and push"
    exit 1
fi

VERSION=$1
PROVIDERS="provider-family-gcp provider-gcp-secretmanager provider-gcp-storage provider-gcp-cloudplatform"

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
