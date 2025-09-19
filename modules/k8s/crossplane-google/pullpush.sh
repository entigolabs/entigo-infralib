

if [ "$1" == "" ]
then
	echo "First parameter is provider version to pull and push"
	exit 1
fi

VERSION=$1

PROVIDERS="provider-family-gcp provider-gcp-secretmanager provider-gcp-storage provider-gcp-cloudplatform" 


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

