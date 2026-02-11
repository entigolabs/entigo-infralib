#!/bin/bash

if [ "$1" == "" ]; then
    VERSION="v2.4.0"
    echo "Defaulting to version $VERSION"
else
    VERSION=$1
fi

PROVIDERS="provider-family-aws provider-aws-elasticache provider-aws-ram provider-aws-networkfirewall provider-aws-ec2 provider-aws-route53 provider-aws-ecr provider-aws-sqs provider-aws-s3 provider-aws-iam provider-aws-vpc provider-aws-secretsmanager provider-aws-rds provider-aws-kms provider-aws-eks provider-aws-kafka provider-aws-ssoadmin provider-aws-accessanalyzer provider-aws-account provider-aws-acm provider-aws-acmpca provider-aws-amp provider-aws-amplify provider-aws-apigateway provider-aws-apigatewayv2 provider-aws-appautoscaling provider-aws-appconfig provider-aws-appflow provider-aws-appintegrations provider-aws-applicationinsights provider-aws-appmesh provider-aws-apprunner provider-aws-appstream provider-aws-appsync provider-aws-athena provider-aws-autoscaling provider-aws-autoscalingplans provider-aws-backup provider-aws-batch provider-aws-bedrock provider-aws-bedrockagent provider-aws-budgets provider-aws-ce provider-aws-chime provider-aws-cloud9 provider-aws-cloudcontrol provider-aws-cloudformation provider-aws-cloudfront provider-aws-cloudsearch provider-aws-cloudtrail provider-aws-cloudwatch provider-aws-cloudwatchevents provider-aws-cloudwatchlogs provider-aws-codeartifact provider-aws-codecommit provider-aws-codeguruprofiler provider-aws-codepipeline provider-aws-codestarconnections provider-aws-codestarnotifications provider-aws-cognitoidentity provider-aws-cognitoidp provider-aws-configservice provider-aws-connect provider-aws-cur provider-aws-dataexchange provider-aws-datapipeline provider-aws-datasync provider-aws-dax provider-aws-deploy provider-aws-detective provider-aws-devicefarm provider-aws-directconnect provider-aws-dlm provider-aws-dms provider-aws-docdb provider-aws-ds provider-aws-dsql provider-aws-dynamodb provider-aws-ecrpublic provider-aws-ecs provider-aws-efs provider-aws-elasticbeanstalk provider-aws-elasticsearch provider-aws-elastictranscoder provider-aws-elb provider-aws-elbv2 provider-aws-emr provider-aws-emrserverless provider-aws-evidently provider-aws-firehose provider-aws-fis provider-aws-fsx provider-aws-gamelift provider-aws-glacier provider-aws-globalaccelerator provider-aws-glue provider-aws-grafana provider-aws-guardduty provider-aws-identitystore provider-aws-imagebuilder provider-aws-inspector provider-aws-inspector2 provider-aws-iot provider-aws-ivs provider-aws-kafkaconnect provider-aws-kendra provider-aws-keyspaces provider-aws-kinesis provider-aws-kinesisanalytics provider-aws-kinesisanalyticsv2 provider-aws-kinesisvideo provider-aws-lakeformation provider-aws-lambda provider-aws-lexmodels provider-aws-licensemanager provider-aws-lightsail provider-aws-location provider-aws-macie2 provider-aws-mediaconvert provider-aws-medialive provider-aws-mediapackage provider-aws-mediastore provider-aws-memorydb provider-aws-mq provider-aws-mwaa provider-aws-neptune provider-aws-networkmanager provider-aws-opensearch provider-aws-opensearchserverless provider-aws-organizations provider-aws-osis provider-aws-pinpoint provider-aws-pipes provider-aws-qldb provider-aws-quicksight provider-aws-redshift provider-aws-redshiftserverless provider-aws-resourcegroups provider-aws-rolesanywhere provider-aws-route53recoverycontrolconfig provider-aws-route53recoveryreadiness provider-aws-route53resolver provider-aws-rum provider-aws-s3control provider-aws-sagemaker provider-aws-scheduler provider-aws-schemas provider-aws-securityhub provider-aws-serverlessrepo provider-aws-servicecatalog provider-aws-servicediscovery provider-aws-servicequotas provider-aws-ses provider-aws-sesv2 provider-aws-sfn provider-aws-signer provider-aws-sns provider-aws-ssm provider-aws-swf provider-aws-timestreamwrite provider-aws-transcribe provider-aws-transfer provider-aws-verifiedaccess provider-aws-vpclattice provider-aws-waf provider-aws-wafregional provider-aws-wafv2 provider-aws-workspaces provider-aws-xray"

for provider in $PROVIDERS; do
    SOURCE="xpkg.upbound.io/upbound/$provider:$VERSION"
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


