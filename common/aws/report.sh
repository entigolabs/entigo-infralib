#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH/../..



for line in $(for verfile in $(find modules/aws/ -name versions.tf | sort)
do
  versionfound=`awk -v keyword="source" '$0 ~ " " keyword { source=$3; gsub(" ", "", source); while($1 != "}") { if($1 == "version") print source "=" $3; getline; } }' $verfile | tr -d '\"'`
  echo $versionfound
done)
do
echo $line

done | sort | uniq | while read line
do
  namespace=$(echo $line | cut -d"/" -f1)
  name=$(echo $line | cut -d"/" -f2 | cut -d"=" -f1)
  currentversion=$(echo $line | cut -d"=" -f2)
  registry=`curl -s "https://registry.terraform.io/v2/providers/$namespace/$name?include=latest-version&name=$name&namespace=$namespace"`
  versionid=`echo $registry | jq -r '.data.relationships["latest-version"].data.id'`
  latestversion=`echo $registry | jq -r --arg id "$versionid" '.included[] | select(.id == $id) | .attributes.version'`
  if [ "$currentversion" != "$latestversion" ]
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
  registry=`curl -s "https://registry.terraform.io/v2/modules/$name?include=latest-version"`
  
  versionid=`echo $registry | jq -r '.data.relationships["latest-version"].data.id'`
  latestversion=`echo $registry | jq -r --arg id "$versionid" '.included[] | select(.id == $id) | .attributes.version'`
  if [ "$currentversion" != "$latestversion" ]
  then
     echo "$name newer version $latestversion, current $currentversion"
  fi
done | sort | uniq
