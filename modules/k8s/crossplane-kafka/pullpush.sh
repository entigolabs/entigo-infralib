#!/bin/bash

if [ "$1" == "" ]; then
    VERSION="v1.0.1"
    echo "Defaulting to version $VERSION"
else
    VERSION=$1
fi

PROVIDERS="provider-kafka"

for provider in $PROVIDERS; do
    SOURCE="xpkg.upbound.io/crossplane-contrib/$provider:$VERSION"
    DEST="entigolabs/$provider:$VERSION"
    if docker manifest inspect $DEST > /dev/null 2>&1; then
        echo "Skipping $provider - already exists at $DEST"
        continue
    fi
    echo "Copying $SOURCE to $DEST"
    docker buildx imagetools create --tag $DEST $SOURCE

    if [ $? -ne 0 ]; then
        echo "Failed to copy $SOURCE to $DEST"
        exit 2
    fi
done

echo "All providers copied successfully"
