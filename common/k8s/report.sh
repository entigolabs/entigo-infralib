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
        # Strip oci:// prefix, build full image ref
        registry_url="${url#oci://}"
        ref="$registry_url/$name"
        registry_host=$(echo "$ref" | cut -d'/' -f1)
        image_path=$(echo "$ref" | cut -d'/' -f2-)

        # Get anonymous bearer token
        if [[ "$registry_host" == "ghcr.io" ]]; then
          token=$(curl -s "https://ghcr.io/token?scope=repository:${image_path}:pull" | jq -r '.token')
        elif [[ "$registry_host" == "public.ecr.aws" ]]; then
          # Note trailing slash — ECR public redirects without it
          token=$(curl -s "https://public.ecr.aws/token/?scope=repository:${image_path}:pull&service=public.ecr.aws" | jq -r '.token')
        fi

        # Fetch all tags, following pagination
        all_tags=""
        next_url="https://${registry_host}/v2/${image_path}/tags/list?n=1000"
        while [ -n "$next_url" ]; do
          response=$(curl -sI -H "Authorization: Bearer $token" "$next_url")
          body=$(curl -s -H "Authorization: Bearer $token" "$next_url")
          all_tags="$all_tags $(echo "$body" | jq -r '.tags[]?' 2>/dev/null)"
          # Check for Link header for next page
          next_url=$(echo "$response" | grep -i '^link:' | sed 's/.*<\(.*\)>.*/\1/' | grep -v '^$')
          # Make relative URLs absolute
          if [[ "$next_url" == /v2/* ]]; then
            next_url="https://${registry_host}${next_url}"
          fi
        done

        latest=$(echo "$all_tags" | tr ' ' '\n' | sed 's/^v//' | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1)

      else
        helm repo add $name $url > /dev/null
        latest=$(helm search repo -r "\v$name/$name\v" --output json | jq -r '.[0].version')
      fi
      if [ "$latest" == "" -o "$latest" == "null" ]
      then
        echo "$name Chart not found in repo $url ($chart)"
      elif [ "$(echo $latest | sed 's/^v//' | cut -d'.' -f1-3)" != "$(echo $version | sed 's/^v//' | cut -d'.' -f1-3)" ]
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
