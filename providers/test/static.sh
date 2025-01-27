#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/.."
cd $SCRIPTPATH || exit 1
pwd
source ../../common/generate_config.sh

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
    versionfound=`awk -v keyword="$providername" '$0 ~ " " keyword { getline; while($1 != "}") { if($1 == "version") print $3; getline; } }' $verfile | tr -d '\"'`
    if [ "$versionfound" != "" ]
    then
      echo "Found $providername version $versionfound in $verfile"
      if [ "$lastversion" == "" ]
      then
        lastversion=$versionfound
      elif [ "$lastversion" != "$versionfound" ]
      then
        echo "Version mismatch for $providername $lastversion != $versionfound in $verfile"
        exit 1
      fi
    fi
  done
  if [ "$providername" == "helmaws" -o "$providername" == "helmgoogle" ]
  then
    modulename="helm"
  else
    modulename="$providername"
  fi
  
  awk -v providername="$providername" -v modulename="$modulename" -v lastversion="$lastversion" '/required_providers {/ { print; print "    " providername " = {\n      source  = \"hashicorp/" modulename "\"\n      version = \"" lastversion "\"\n    }"; next }1' test_base.tf > tmp && mv tmp test_base.tf

done

if [ -d "tmp_tf" ]
then
	rm -rf tmp_tf
fi
mkdir tmp_tf
cp *.tf tmp_tf/
rm tmp_tf/base.tf


docker run --rm -v "$(pwd)/tmp_tf":"/data" $TFLINT_IMAGE
