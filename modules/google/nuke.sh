#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH || exit 1


if [ "$GITHUB_ACTION" != "" ]
then
  mkdir -p $(echo ~)/.config/gcloud 
  echo ${GOOGLE_CREDENTIALS} > $(echo ~)/.config/gcloud/application_default_credentials.json
fi


gcloud -q config set project "entigo-infralib" || exit 1


gsutil ls | while read line
do
  echo "gsutil rm -r $line"

done

gcloud -q "compute" "firewall-rules" list --uri | while read line
do
  echo "gcloud 'compute' 'firewall-rules' delete --project entigo-infralib -q $line"
done

gcloud -q "compute" "networks" list --uri | while read line
do
  echo "gcloud 'compute' 'networks' delete --project entigo-infralib -q $line"
done

gcloud -q "compute" "routes" list --uri | while read line
do
  echo "gcloud 'compute' 'routes' delete --project entigo-infralib -q $line"
done

gcloud -q "compute" "networks" list --uri | while read line
do
  echo "gcloud 'compute' 'networks' delete --project entigo-infralib -q $line"
done

gcloud "secrets" list --uri | while read line
do
  echo "gcloud 'secrets' delete --project entigo-infralib -q $line"

done
