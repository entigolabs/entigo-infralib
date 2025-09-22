

if [ "$1" == "" ]
then
	echo "First parameter is provider version to pull and push"
	exit 1
fi

VERSION=$1

PROVIDERS="provider-family-aws provider-aws-elasticache provider-aws-ram provider-aws-networkfirewall provider-aws-ec2 provider-aws-route53 provider-aws-ecr provider-aws-sqs provider-aws-s3 provider-aws-iam provider-aws-vpc provider-aws-secretsmanager provider-aws-rds provider-aws-kms provider-aws-eks"


for provider in $PROVIDERS
do
  docker pull xpkg.upbound.io/upbound/$provider:$VERSION
  if [ $? -ne 0 ]
  then
	  echo "Failed to pull xpkg.upbound.io/upbound/$provider:$VERSION"
	  exit 2
  fi
done

for provider in $PROVIDERS
do
  docker tag xpkg.upbound.io/upbound/$provider:$VERSION entigolabs/$provider:$VERSION
  if [ $? -ne 0 ]
  then
          echo "Failed to tag xpkg.upbound.io/upbound/$provider:$VERSION to entigolabs/$provider:$VERSION"
          exit 3
  fi
  docker push entigolabs/$provider:$VERSION
  if [ $? -ne 0 ]
  then
          echo "Failed to push entigolabs/$provider:$VERSION"
          exit 4
  fi
done

