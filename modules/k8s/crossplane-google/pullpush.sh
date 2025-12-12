#!/bin/bash

if [ "$1" == "" ]; then
    VERSION="v2.3.0"
    echo "Defaulting to version $VERSION"
else
    VERSION=$1
fi

PROVIDERS="provider-family-gcp provider-gcp-secretmanager provider-gcp-storage provider-gcp-cloudplatform provider-gcp-accesscontextmanager provider-gcp-activedirectory provider-gcp-alloydb provider-gcp-apigee provider-gcp-appengine provider-gcp-artifact provider-gcp-beyondcorp provider-gcp-bigquery provider-gcp-bigtable provider-gcp-binaryauthorization provider-gcp-certificatemanager provider-gcp-cloud provider-gcp-cloudbuild provider-gcp-cloudfunctions provider-gcp-cloudfunctions2 provider-gcp-cloudquotas provider-gcp-cloudrun provider-gcp-cloudscheduler provider-gcp-cloudtasks provider-gcp-composer provider-gcp-compute provider-gcp-container provider-gcp-containeranalysis provider-gcp-containerattached provider-gcp-containeraws provider-gcp-containerazure provider-gcp-datacatalog provider-gcp-dataflow provider-gcp-datafusion provider-gcp-datalossprevention provider-gcp-dataplex provider-gcp-dataproc provider-gcp-datastream provider-gcp-developerconnect provider-gcp-dialogflowcx provider-gcp-dns provider-gcp-documentai provider-gcp-essentialcontacts provider-gcp-eventarc provider-gcp-filestore provider-gcp-firebaserules provider-gcp-gemini provider-gcp-gke provider-gcp-gkehub provider-gcp-healthcare provider-gcp-iam provider-gcp-iap provider-gcp-identityplatform provider-gcp-kms provider-gcp-logging provider-gcp-memcache provider-gcp-mlengine provider-gcp-monitoring provider-gcp-networkconnectivity provider-gcp-networkmanagement provider-gcp-networksecurity provider-gcp-networkservices provider-gcp-notebooks provider-gcp-orgpolicy provider-gcp-osconfig provider-gcp-oslogin provider-gcp-privateca provider-gcp-pubsub provider-gcp-redis provider-gcp-servicenetworking provider-gcp-sourcerepo provider-gcp-spanner provider-gcp-sql provider-gcp-storagetransfer provider-gcp-tags provider-gcp-tpu provider-gcp-vertexai provider-gcp-vpcaccess provider-gcp-workflows"

for provider in $PROVIDERS; do
    SOURCE="xpkg.upbound.io/upbound/$provider:$VERSION"
    DEST="entigolabs/$provider:$VERSION"
    # Check if destination tag already exists
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
