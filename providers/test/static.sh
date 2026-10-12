#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/.."
cd $SCRIPTPATH || exit 1
pwd
source ../common/generate_config.sh

export DOCKER_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
  export DOCKER_OPTS="-it"
fi


cp base.tf test_base.tf
for line in `ls -1 *.tf | grep -ve"base.tf\|test_base.tf"`
do
  echo "Version unity check of $line"
  lastversion=""
  providername=`echo $line | cut -d"." -f1`

  for verfile in `find ../modules/ -name versions.tf`
  do
    while read -r versionfound
    do
      if [ "$versionfound" == "" ]
      then
        continue
      fi
      echo "Found $providername version $versionfound in $verfile"
      if ! [[ "$versionfound" =~ ^[0-9]+(\.[0-9]+)*$ ]]
      then
        echo "Skipping $verfile, $providername version $versionfound is a constraint and not a pinned version"
        continue
      fi
      if [ "$lastversion" == "" ]
      then
        lastversion=$versionfound
      elif [ "$lastversion" != "$versionfound" ]
      then
        echo "Version mismatch for $providername $lastversion != $versionfound in $verfile"
        exit 1
      fi
    done < <(awk -v keyword="$providername" '$1 == keyword && $2 == "=" { while(getline > 0 && $1 != "}") { if($1 == "version") { sub(/^[[:space:]]*version[[:space:]]*=[[:space:]]*/, ""); sub(/[[:space:]]*$/, ""); gsub(/"/, ""); print; } } }' $verfile)
  done
  if [ "$providername" == "helmaws" -o "$providername" == "helmgoogle" ]
  then
    modulename="helm"
  else
    modulename="$providername"
  fi

  if [ "$providername" == "oci" ]
  then
    sourceorg="oracle"
  else
    sourceorg="hashicorp"
  fi

  awk -v providername="$providername" -v modulename="$modulename" -v sourceorg="$sourceorg" -v lastversion="$lastversion" '/required_providers {/ { print; print "    " providername " = {\n      source  = \"" sourceorg "/" modulename "\"\n      version = \"" lastversion "\"\n    }"; next }1' test_base.tf > tmp && mv tmp test_base.tf

done

if [ -d "tmp_tf" ]
then
	rm -rf tmp_tf
fi
mkdir tmp_tf
cp *.tf tmp_tf/
rm tmp_tf/base.tf


docker run --rm -v "$(pwd)/tmp_tf":"/data" $TFLINT_IMAGE
