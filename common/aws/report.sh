#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH/../..



for line in $(for verfile in $(find modules/aws/ -name versions.tf | sort)
do
  versionfound=`awk -v keyword="source" '$0 ~ " " keyword { source=$3; gsub(" ", "", source); while($1 != "}") { if($1 == "version") { v=""; for(i=3;i<=NF;i++) v=v$i; print source "=" v }; getline; } }' $verfile | tr -d '\"'`
  echo $versionfound
done)
do
echo $line

done | sort | uniq | while read line
do
  namespace=$(echo $line | cut -d"/" -f1)
  name=$(echo $line | cut -d"/" -f2 | cut -d"=" -f1)
  currentversion=$(echo $line | cut -d"=" -f2-)
  # OpenTofu registry, standard registry protocol. Unlike the Terraform
  # registry's v2 API it returns every version rather than a latest pointer, so
  # drop prereleases and take the highest.
  latestversion=`curl -s "https://registry.opentofu.org/v1/providers/$namespace/$name/versions" | jq -r '.versions[]?.version' | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1`
  if [ "$latestversion" == "" ]
  then
     echo "$namespace/$name provider not found in https://registry.opentofu.org/v1/providers/$namespace/$name/versions"
  elif [[ "$currentversion" != "$latestversion" && "$currentversion" != ">="* ]]
  then
     echo "$namespace/$name newer version $latestversion, current $currentversion"
  fi
done


for line in $(for verfile in $(find modules/aws/ -name main.tf | sort)
do
  versionfound=`awk -v keyword="source" '$0 ~ " " keyword { source=$3; gsub(" ", "", source); while($1 != "}") { if($1 == "version") print source "=" $3; getline; } }' $verfile | tr -d '\"'`
  echo $versionfound
done)
do
echo "$line"

done | sort | uniq | while read line
do
  name=$(echo $line | cut -d"=" -f1)
  name=${name%%//*}
  currentversion=$(echo $line | cut -d"=" -f2)
  # $name is already namespace/name/system, which is what the module endpoint takes
  latestversion=`curl -s "https://registry.opentofu.org/v1/modules/$name/versions" | jq -r '.modules[0].versions[]?.version' | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1`
  if [ "$latestversion" == "" ]
  then
     echo "$name module not found in https://registry.opentofu.org/v1/modules/$name/versions"
  elif [ "$currentversion" != "$latestversion" ]
  then
     echo "$name newer version $latestversion, current $currentversion"
  fi
done | sort | uniq


AL2023_CURRENT=$(cat modules/aws/eks/main.tf | grep '"AL2023_x86_64_STANDARD"' | cut -d'"' -f4)
AL2023_LATEST=$(aws ssm get-parameter   --name "/aws/service/eks/optimized-ami/1.35/amazon-linux-2023/arm64/standard/recommended/release_version"   --region eu-north-1 --query "Parameter.Value" --output text)
if [ "$AL2023_CURRENT" != "$AL2023_LATEST" ]
then
  echo "EKS AL2023_* newer version $AL2023_LATEST, current $AL2023_CURRENT"
fi

BOTTLE_CURRENT=$(cat modules/aws/eks/main.tf | grep '"BOTTLEROCKET_x86_64"' | cut -d'"' -f4)
BOTTLE_LATEST=$(aws ssm get-parameter   --name "/aws/service/bottlerocket/aws-k8s-1.35/x86_64/latest/image_version"   --region eu-north-1 --query "Parameter.Value" --output text)
if [ "$BOTTLE_CURRENT" != "$BOTTLE_LATEST" ]
then
  echo "EKS AL2023_* newer version $BOTTLE_LATEST, current $BOTTLE_CURRENT"
fi
