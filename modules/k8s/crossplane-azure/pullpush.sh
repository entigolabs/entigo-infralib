#!/bin/bash

if [ "$1" == "" ]; then
    VERSION="v2.7.2"
    echo "Defaulting to version $VERSION"
else
    VERSION=$1
fi

PROVIDERS="provider-family-azure provider-azure-alertsmanagement provider-azure-analysisservices provider-azure-apimanagement provider-azure-appconfiguration provider-azure-appplatform provider-azure-attestation provider-azure-authorization provider-azure-automation provider-azure-azurestackhci provider-azure-botservice provider-azure-cache provider-azure-cdn provider-azure-certificateregistration provider-azure-cognitiveservices provider-azure-communication provider-azure-compute provider-azure-confidentialledger provider-azure-consumption provider-azure-containerapp provider-azure-containerregistry provider-azure-containerservice provider-azure-cosmosdb provider-azure-costmanagement provider-azure-customproviders provider-azure-dashboard provider-azure-databoxedge provider-azure-databricks provider-azure-datafactory provider-azure-datamigration provider-azure-dataprotection provider-azure-datashare provider-azure-dbformysql provider-azure-dbforpostgresql provider-azure-desktopvirtualization provider-azure-devcenter provider-azure-devices provider-azure-deviceupdate provider-azure-devopsinfrastructure provider-azure-devtestlab provider-azure-digitaltwins provider-azure-elastic provider-azure-eventgrid provider-azure-eventhub provider-azure-fluidrelay provider-azure-guestconfiguration provider-azure-hdinsight provider-azure-healthbot provider-azure-healthcareapis provider-azure-insights provider-azure-iotcentral provider-azure-keyvault provider-azure-kusto provider-azure-loadtestservice provider-azure-logic provider-azure-machinelearningservices provider-azure-maintenance provider-azure-managedidentity provider-azure-management provider-azure-maps provider-azure-marketplaceordering provider-azure-netapp provider-azure-network provider-azure-notificationhubs provider-azure-operationalinsights provider-azure-operationsmanagement provider-azure-oracle provider-azure-orbital provider-azure-policyinsights provider-azure-portal provider-azure-powerbidedicated provider-azure-purview provider-azure-recoveryservices provider-azure-relay provider-azure-resources provider-azure-search provider-azure-security provider-azure-securityinsights provider-azure-servicebus provider-azure-servicefabric provider-azure-servicelinker provider-azure-servicenetworking provider-azure-signalrservice provider-azure-solutions provider-azure-spring provider-azure-sql provider-azure-storage provider-azure-storagecache provider-azure-storagesync provider-azure-streamanalytics provider-azure-synapse provider-azure-web"

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
