#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH/../..


for chart in $(find modules/k8s/ -name Chart.yaml | sort)
do
  yq -r '.dependencies[] | "\(.name) \(.version) \(.repository)"' $chart | while read dep
  do
    if [ "$dep" != "" ]
    then
      name=$(echo $dep | awk '{print $1}')
      version=$(echo $dep | awk '{print $2}')
      url=$(echo $dep | awk '{print $3}') 
      if [[ "$url" == oci* ]]
      then
        #no support for oci yet...
        latest=$version
      else
        helm repo add $name $url > /dev/null
        latest=$(helm search repo -r "\v$name/$name\v" --output json | jq -r '.[0].version')
      fi
      if [ "$latest" == "" -o "$latest" == "null" ]
      then
        echo "$name Chart not found in repo $url ($chart)"
      elif [ "$(echo $latest | cut -d'.' -f1-2)" != "$(echo $version | cut -d'.' -f1-2)" ]
      then
        echo "$name newer version $latest, current $version ($chart)"
      fi
    fi
  done

done
