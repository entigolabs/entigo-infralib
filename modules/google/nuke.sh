#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH || exit 1


if [ "$GITHUB_ACTION" != "" ]
then
  mkdir -p $(echo ~)/.config/gcloud 
  echo ${GOOGLE_CREDENTIALS} > $(echo ~)/credentials.json
  gcloud -q auth activate-service-account --key-file $(echo ~)/credentials.json || exit 1
fi

gcloud -q config set project "entigo-infralib" || exit 1

gsutil ls | while read line
do
  gsutil rm -r $line
done

gcloud -q "compute" "firewall-rules" list --uri | while read line
do
  gcloud 'compute' 'firewall-rules' delete --project entigo-infralib -q $line
done

gcloud container clusters list --uri | while read line
do
  gcloud 'container' 'clusters' delete --project entigo-infralib --timeout 3600  -q $line
done

gcloud run jobs list --uri | while read line
do
  gcloud 'run' 'jobs' delete --project entigo-infralib --region europe-north1 -q $line
done

gcloud compute routers list --uri | while read line
do
  gcloud 'compute' 'routers' delete --project entigo-infralib -q $line
done

gcloud -q "compute" "networks" "subnets" list --uri | while read line
do
  gcloud 'compute' "networks" "subnets" delete --project entigo-infralib -q $line
done

gcloud -q "compute" "routes" list --uri | while read line
do
  gcloud 'compute' 'routes' delete --project entigo-infralib -q $line
done

gcloud -q "compute" "networks" list --uri | while read line
do
  gcloud 'compute' 'networks' delete --project entigo-infralib -q $line
done

gcloud "secrets" list --uri | while read line
do
  gcloud 'secrets' delete --project entigo-infralib -q $line
done

gcloud dns managed-zones list --uri | grep -ve"gcp-infralib-entigo-io" | while read line
do
  gcloud 'dns' 'managed-zones' delete --project entigo-infralib -q $line
done

gcloud iam service-accounts list --format='value(email)' | grep -vE '175436099636-compute@developer.gserviceaccount.com|infralib-agent@entigo-infralib.iam.gserviceaccount.com|github@entigo-infralib.iam.gserviceaccount.com' | while read line
do
  gcloud 'iam' 'service-accounts' delete --project entigo-infralib -q $line
done