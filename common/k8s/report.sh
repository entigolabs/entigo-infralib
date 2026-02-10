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

for providerfile in $(find modules/k8s/ -name provider.yaml | sort)
do
  cat $providerfile | grep xpkg.upbound.io | grep -ve"\$provider" | cut -d"/" -f2- | while read provider
  do
    old_version=$(echo $provider | cut -d":" -f2)
    provider_path=$(echo $provider | cut -d":" -f1)
    latest_version=$(curl -s https://marketplace.upbound.io/providers/$provider_path | cut -d"/" -f5)
    if [ "$latest_version" == "" ]
    then
      echo "$provider_path crossplane provider package latest version https://marketplace.upbound.io/providers/$provider_path not found"
    elif [ "$old_version" != "$latest_version" ]
    then
      echo "$provider_path newer version $latest_version, current $old_version ($providerfile)"
    fi
  done
done
