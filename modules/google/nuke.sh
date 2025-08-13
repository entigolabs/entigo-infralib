#!/bin/bash
SCRIPTPATH="$(
  cd -- "$(dirname "$0")" >/dev/null 2>&1
  pwd -P
)"
cd $SCRIPTPATH || exit 1

if [ "$GITHUB_ACTION" != "" ]; then
  mkdir -p $(echo ~)/.config/gcloud
  echo ${GOOGLE_CREDENTIALS} >$(echo ~)/credentials.json
  gcloud -q auth activate-service-account --key-file $(echo ~)/credentials.json || exit 1
fi

export GOOGLE_PROJECT="entigo-infralib2"
export GOOGLE_REGION="europe-north1"

gcloud -q config set project $GOOGLE_PROJECT || exit 1
gcloud -q config set compute/region $GOOGLE_REGION || exit 1

PIDS=""
for line in $(gcloud -q deploy delivery-pipelines list --project $GOOGLE_PROJECT --region $GOOGLE_REGION --uri); do
  gcloud deploy delivery-pipelines delete --project $GOOGLE_PROJECT --region $GOOGLE_REGION --force -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete delivery pipelines. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud deploy targets list --project $GOOGLE_PROJECT --region $GOOGLE_REGION | grep "name:" | awk '{print $2}'); do
  gcloud deploy targets delete --project $GOOGLE_PROJECT --region $GOOGLE_REGION -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete deploy targets. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "firewall-rules" list --uri); do
  gcloud 'compute' "firewall-rules" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete firewall-rules. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q run jobs list --uri); do
  gcloud run jobs delete --project $GOOGLE_PROJECT --region $GOOGLE_REGION -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete run jobs. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q dns policies list --uri); do
  gcloud dns policies update --project $GOOGLE_PROJECT --networks="" -q $line
  gcloud dns policies delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete DNS policies. $FAIL"
  exit 1
fi

delete_cluster() {
  local cluster_uri=$1
  local max_retries=20
  local sleep_time=60

  for ((i = 1; i <= $max_retries; i++)); do
    echo "Attempt $i to delete cluster: $cluster_uri"
    gcloud container clusters delete --project $GOOGLE_PROJECT --region $GOOGLE_REGION --timeout 3600 -q $cluster_uri
    if [ $? -eq 0 ]; then
      echo "Cluster $cluster_uri deleted successfully."
      return 0
    else
      if [ $i -lt $max_retries ]; then
        echo "Failed to delete cluster $cluster_uri. Retrying in $sleep_time seconds..."
        sleep $sleep_time
      else
        echo "Failed to delete cluster $cluster_uri after $max_retries attempts."
        return 1
      fi
    fi
  done
}
PIDS=""
FAIL=0
for cluster_uri in $(gcloud container clusters list --uri); do
  delete_cluster "$cluster_uri" &
  PIDS="$PIDS $!"
done
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete container clusters. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "forwarding-rules" list --uri); do
  gcloud "compute" "forwarding-rules" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete forwarding-rules. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "target-http-proxies" list --uri); do
  gcloud "compute" "target-http-proxies" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete target-http-proxies. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "target-https-proxies" list --uri); do
  gcloud "compute" "target-https-proxies" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete target-https-proxies. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "url-maps" list --uri); do
  gcloud "compute" "url-maps" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete url-maps. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "backend-services" list --uri); do
  gcloud "compute" "backend-services" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete backend-services. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "network-endpoint-groups" list --uri); do
  gcloud "compute" "network-endpoint-groups" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete network-endpoint-groups. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "routers" list --uri); do
  gcloud "compute" "routers" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete routers. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "addresses" list --uri); do
  gcloud 'compute' "addresses" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete addresses. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "networks" "subnets" list --uri); do
  gcloud 'compute' "networks" "subnets" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete subnets. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "routes" list --uri); do
  gcloud "compute" "routes" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete routes. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "networks" list --uri); do
  gcloud 'compute' 'networks' delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete compute networks $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q secrets list --uri | grep -ve"argocd-biz_clientSecret\|argocd-pri_clientSecret\|gar-proxy-"); do
  gcloud secrets delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete secrets. $FAIL"
  exit 1
fi

gcloud -q certificate-manager maps list --uri | while read -r MAP; do
  gcloud -q certificate-manager maps entries list --uri --map=$MAP | while read -r ENTRY; do
    gcloud certificate-manager maps entries delete --project $GOOGLE_PROJECT -q $ENTRY
  done

  MAX_RETRIES=20
  RETRY_DELAY=60
  SUCCESS=false

  for ((i = 1; i <= MAX_RETRIES; i++)); do
    echo "Attempt $i of $MAX_RETRIES to delete map $MAP..."
    if gcloud certificate-manager maps delete --project $GOOGLE_PROJECT -q $MAP; then
      SUCCESS=true
      break
    fi
    echo "Retrying in $RETRY_DELAY seconds..."
    sleep $RETRY_DELAY
  done

  if [ "$SUCCESS" = true ]; then
    echo "Successfully deleted map $MAP."
  else
    echo "Failed to delete map $MAP after $MAX_RETRIES attempts."
  fi
done

PIDS=""
for line in $(gcloud -q "certificate-manager" "certificates" list --uri); do
  gcloud "certificate-manager" "certificates" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete certificates. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "certificate-manager" "certificates" list --location $GOOGLE_REGION --uri); do
  gcloud "certificate-manager" "certificates" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete certificates. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "certificate-manager" "dns-authorizations" list --uri); do
  gcloud "certificate-manager" "dns-authorizations" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete dns-authorizations. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "certificate-manager" "dns-authorizations" list --location $GOOGLE_REGION --uri); do
  gcloud "certificate-manager" "dns-authorizations" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete dns-authorizations. $FAIL"
  exit 1
fi

gcloud dns managed-zones list --format="get(name)" | grep -vEx "gcp-infralib-entigo-io|agent-gcp-infralib-entigo-io" | while read -r ZONE_NAME; do
  gcloud dns record-sets list --zone=$ZONE_NAME --format="get(name,type)" | while read -r RECORD_NAME TYPE; do
    OUTPUT=$(gcloud dns record-sets delete --zone=$ZONE_NAME --type=$TYPE --project=$GOOGLE_PROJECT -q $RECORD_NAME 2>&1)
    if ! echo "$OUTPUT" | grep -q "HTTPError 400: The resource record set .* is invalid because a zone must contain exactly one resource record set of type .* at the apex."; then
      echo "$OUTPUT"
    fi
  done
  gcloud dns managed-zones delete --project $GOOGLE_PROJECT -q $ZONE_NAME
done

gcloud dns record-sets list --zone=gcp-infralib-entigo-io --format="get(name,type)" | grep -ve "^gcp.infralib.entigo.io.\|^agent.gcp.infralib.entigo.io." | while read -r RECORD_NAME TYPE; do
  gcloud dns record-sets delete --type=$TYPE --zone=gcp-infralib-entigo-io --project $GOOGLE_PROJECT -q $RECORD_NAME
done

PIDS=""
for line in $(gcloud -q "compute" "ssl-certificates" list --uri); do
  gcloud 'compute' "ssl-certificates" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete ssl-certificates. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "health-checks" list --uri); do
  gcloud 'compute' "health-checks" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete health-checks. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud -q "compute" "disks" list --uri); do
  gcloud 'compute' "disks" delete --project $GOOGLE_PROJECT -q $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete disks. $FAIL"
  exit 1
fi

gcloud iam service-accounts list --format='value(email)' | grep -vE 'compute@developer.gserviceaccount.com|infralib-agent|github' | while read line; do
  gcloud 'iam' 'service-accounts' delete --project $GOOGLE_PROJECT -q $line
done

PIDS=""
for line in $(gcloud -q artifacts repositories list --format='value(name)' --location="$GOOGLE_REGION" --project="$GOOGLE_PROJECT" | grep -vE 'entigolabs'); do
  gcloud artifacts repositories delete "$line" --project="$GOOGLE_PROJECT" --location="$GOOGLE_REGION" -q &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete artifact repositories. $FAIL"
  exit 1
fi

PIDS=""
for line in $(gcloud storage ls); do
  gcloud storage rm --recursive $line &
  PIDS="$PIDS $!"
done
FAIL=0
for p in $PIDS; do
  wait $p || let "FAIL+=1"
  echo $p $FAIL
done
if [ "$FAIL" -ne 0 ]; then
  echo "FAILED to delete storage buckets. $FAIL"
  exit 1
fi
