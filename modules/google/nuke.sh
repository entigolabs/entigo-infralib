#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH || exit 1


if [ "$GITHUB_ACTION" != "" ]
then
  mkdir -p $(echo ~)/.config/gcloud 
  echo ${GOOGLE_CREDENTIALS} > $(echo ~)/credentials.json
  gcloud -q auth activate-service-account --key-file $(echo ~)/credentials.json || exit 1
fi

gcloud -q config set project "entigo-infralib2" || exit 1
gcloud -q config set compute/region "europe-north1" || exit 1

gsutil ls | while read line
do
  gsutil rm -r $line
done

gcloud deploy delivery-pipelines list --project entigo-infralib2 --region europe-north1 --uri | while read line
do
  gcloud deploy delivery-pipelines delete --project entigo-infralib2 --region europe-north1 --force -q $line
done

gcloud deploy targets list --project entigo-infralib2 --region europe-north1 --uri | while read line
do
  gcloud deploy targets delete --project entigo-infralib2 --region europe-north1 --force -q $line
done


gcloud -q "compute" "firewall-rules" list --uri | while read line
do
  gcloud 'compute' 'firewall-rules' delete --project entigo-infralib2 -q $line
done

PIDS=""
for line in $(gcloud container clusters list --uri)
do
  gcloud 'container' 'clusters' delete --project entigo-infralib2 --region europe-north1 --timeout 3600 -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
    wait $p || let "FAIL+=1"
    echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]
then
  echo "FAILED to delete container clusters. $FAIL"
  exit 1
fi


gcloud run jobs list --uri | while read line
do
  gcloud 'run' 'jobs' delete --project entigo-infralib2 --region europe-north1 -q $line
done

gcloud compute network-endpoint-groups list --uri | while read line
do
  gcloud 'compute' 'network-endpoint-groups' delete --project entigo-infralib2 -q $line
done


gcloud compute routers list --uri | while read line
do
  gcloud 'compute' 'routers' delete --project entigo-infralib2 -q $line
done

PIDS=""
for line in $(gcloud -q "compute" "networks" "subnets" list --uri)
do
  gcloud 'compute' "networks" "subnets" delete --project entigo-infralib2 -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
    wait $p || let "FAIL+=1"
    echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]
then
  echo "FAILED to delete container clusters. $FAIL"
  exit 1
fi

gcloud -q "compute" "routes" list --uri | while read line
do
  gcloud 'compute' 'routes' delete --project entigo-infralib2 -q $line
done


PIDS=""
for line in $(gcloud -q "compute" "networks" list --uri)
do
  gcloud 'compute' 'networks' delete --project entigo-infralib2 -q $line &
  PIDS="$PIDS $!"
done

FAIL=0
for p in $PIDS; do
    wait $p || let "FAIL+=1"
    echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]
then
  echo "FAILED to delete compute networks $FAIL"
  exit 1
fi

gcloud "secrets" list --uri | while read line
do
  gcloud 'secrets' delete --project entigo-infralib2 -q $line
done

gcloud dns managed-zones list --format="get(name)" | grep -vEx "gcp-infralib-entigo-io" | while read -r ZONE_NAME
do
  gcloud dns record-sets list --zone=$ZONE_NAME --format="get(name,type)" | while read -r RECORD_NAME TYPE
  do
    gcloud dns record-sets delete --zone=$ZONE_NAME --type=$TYPE --project entigo-infralib2 -q $RECORD_NAME
  done
  gcloud dns managed-zones delete --project entigo-infralib2 -q $ZONE_NAME
done

gcloud dns record-sets list --zone=gcp-infralib-entigo-io --format="get(name)" | grep -vEx "gcp.infralib.entigo.io." | while read -r RECORD_NAME
do
  gcloud dns record-sets delete --type=NS --zone=gcp-infralib-entigo-io --project entigo-infralib2 -q $RECORD_NAME
done

gcloud compute ssl-certificates list --uri | while read line
do
  gcloud 'compute' 'ssl-certificates' delete --project entigo-infralib2 -q $line
done

gcloud iam service-accounts list --format='value(email)' | grep -vE 'compute@developer.gserviceaccount.com|infralib-agent|github' | while read line
do
  gcloud 'iam' 'service-accounts' delete --project entigo-infralib2 -q $line
done
