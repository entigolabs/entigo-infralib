#!/bin/bash
# Deletes Entigo Infralib Oracle Cloud test resources so provisioning can be re-tested from
# scratch. There's no established "oci-nuke" equivalent of aws-nuke/gcp-nuke, so this is
# hand-rolled against the specific resource types the agent bootstrap + oracle/vpc module
# create (see oracle/{oracle,iam,logging,storage}.go in entigo-infralib-agent for naming).
#
# Compartment-scoped resources (VCNs, buckets, log group, DevOps project, ONS topic) are swept
# unconditionally within ORACLE_COMPARTMENT_ID, on the assumption that it's a dedicated test
# compartment. Tenancy-scoped resources (dynamic groups, policies) and the per-user customer
# secret key are filtered by PREFIX, since those live in a namespace shared across the whole
# tenancy and must not touch other people's resources.
set -uo pipefail
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd "$SCRIPTPATH" || exit 1

if [ "$ORACLE_COMPARTMENT_ID" == "" ]
then
  echo "ERROR: ORACLE_COMPARTMENT_ID must be set."
  exit 1
fi

if [ "$OCI_REGION" == "" ]
then
  echo "Defaulting OCI_REGION to eu-frankfurt-1"
  export OCI_REGION="eu-frankfurt-1"
fi

if [ "$PREFIX" == "" ]
then
  echo "ERROR: PREFIX must be set (used to find state/config buckets, log group, dynamic group/policy)."
  exit 1
fi

OCI="oci --region $OCI_REGION"
COMP="$ORACLE_COMPARTMENT_ID"
TENANCY=$(grep -m1 '^tenancy=' "${OCI_CONFIG_FILE:-$HOME/.oci/config}" | cut -d'=' -f2)
USER_OCID=$(grep -m1 '^user=' "${OCI_CONFIG_FILE:-$HOME/.oci/config}" | cut -d'=' -f2)

echo "Nuking Oracle Cloud test resources"
echo "  compartment: $COMP"
echo "  region:      $OCI_REGION"
echo "  prefix:      $PREFIX"

echo "--- Container Instances ---"
$OCI container-instances container-instance list --compartment-id "$COMP" --all \
  --query "data.items[?\"lifecycle-state\" != 'DELETED' && \"lifecycle-state\" != 'DELETING'].id" 2>/dev/null \
  | jq -r '.[]' | while read -r id; do
    echo "Deleting container instance $id"
    $OCI container-instances container-instance delete --container-instance-id "$id" --force
done

echo "--- DevOps projects ---"
$OCI devops project list --compartment-id "$COMP" --all --query "data.items[].id" 2>/dev/null | jq -r '.[]' | while read -r proj; do
    for pl in $($OCI devops deploy-pipeline list --project-id "$proj" --all --query "data.items[].id" 2>/dev/null | jq -r '.[]'); do
      echo "Deleting deploy pipeline $pl"
      $OCI devops deploy-pipeline delete --deploy-pipeline-id "$pl" --force
    done
    for bp in $($OCI devops build-pipeline list --project-id "$proj" --all --query "data.items[].id" 2>/dev/null | jq -r '.[]'); do
      echo "Deleting build pipeline $bp"
      $OCI devops build-pipeline delete --build-pipeline-id "$bp" --force
    done
    echo "Deleting devops project $proj"
    $OCI devops project delete --project-id "$proj" --force
done

echo "--- ONS notification topics ---"
$OCI ons topic list --compartment-id "$COMP" --all --query "data[].\"topic-id\"" 2>/dev/null | jq -r '.[]' | while read -r topic; do
    echo "Deleting ONS topic $topic"
    $OCI ons topic delete --topic-id "$topic" --force
done

echo "--- Object Storage buckets ---"
NS=$($OCI os ns get --query data --raw-output 2>/dev/null)
for bucket in "${PREFIX}-${OCI_REGION}" "${PREFIX}-${OCI_REGION}-config"; do
    if $OCI os bucket get --namespace "$NS" --bucket-name "$bucket" >/dev/null 2>&1
    then
      # Buckets are created with versioning enabled (see oracle/storage.go), so a plain
      # bulk-delete only removes current versions and leaves the bucket non-empty. Every
      # version, including delete markers, must be deleted explicitly by name+version-id.
      echo "Emptying bucket $bucket (all object versions)"
      $OCI os object list-object-versions --namespace "$NS" --bucket-name "$bucket" --all \
        --query "data[].[name,\"version-id\"]" 2>/dev/null | jq -r '.[] | @tsv' | while IFS=$'\t' read -r name version; do
          echo "Deleting object version $name@$version"
          $OCI os object delete --namespace "$NS" --bucket-name "$bucket" --name "$name" --version-id "$version" --force
      done
      echo "Deleting bucket $bucket"
      $OCI os bucket delete --namespace "$NS" --bucket-name "$bucket" --force
    fi
done

echo "--- Logging ---"
LOG_GROUP_ID=$($OCI logging log-group list --compartment-id "$COMP" --all \
  --query "data[?\"display-name\"=='${PREFIX}-logs'].id | [0]" --raw-output 2>/dev/null)
if [ "$LOG_GROUP_ID" != "" -a "$LOG_GROUP_ID" != "null" ]
then
  for log in $($OCI logging log list --log-group-id "$LOG_GROUP_ID" --all --query "data[].id" 2>/dev/null | jq -r '.[]'); do
    echo "Deleting log $log"
    $OCI logging log delete --log-group-id "$LOG_GROUP_ID" --log-id "$log" --force
  done
  echo "Deleting log group $LOG_GROUP_ID"
  $OCI logging log-group delete --log-group-id "$LOG_GROUP_ID" --force
fi

# Retries a delete command a few times, tolerating the brief eventual-consistency window
# where OCI still reports a dependent resource (e.g. a just-deleted route table) as attached.
retry_delete() {
    local desc="$1"; shift
    local attempt=1
    until "$@" 2>/tmp/nuke_err
    do
      if grep -q "in use\|references\|IncorrectState\|Conflict" /tmp/nuke_err && [ "$attempt" -lt 6 ]
      then
        echo "$desc still referenced, retrying ($attempt/5)..."
        sleep 5
        attempt=$((attempt + 1))
      else
        cat /tmp/nuke_err
        return 1
      fi
    done
    rm -f /tmp/nuke_err
}

echo "--- VCNs (subnets, route tables, gateways) ---"
for vcn in $($OCI network vcn list --compartment-id "$COMP" --all --query "data[].id" 2>/dev/null | jq -r '.[]'); do
    echo "Cleaning VCN $vcn"
    for subnet in $($OCI network subnet list --compartment-id "$COMP" --vcn-id "$vcn" --all --query "data[].id" 2>/dev/null | jq -r '.[]'); do
      echo "Deleting subnet $subnet"
      $OCI network subnet delete --subnet-id "$subnet" --force
    done
    # Deletes every route table, including the VCN's default one - harmless since the whole
    # VCN is being removed anyway, and route rules are what block gateway deletion below.
    for rt in $($OCI network route-table list --compartment-id "$COMP" --vcn-id "$vcn" --all --query "data[].id" 2>/dev/null | jq -r '.[]'); do
      echo "Deleting route table $rt"
      $OCI network route-table delete --rt-id "$rt" --force
    done
    for igw in $($OCI network internet-gateway list --compartment-id "$COMP" --vcn-id "$vcn" --all --query "data[].id" 2>/dev/null | jq -r '.[]'); do
      echo "Deleting internet gateway $igw"
      retry_delete "internet gateway $igw" $OCI network internet-gateway delete --ig-id "$igw" --force
    done
    for nat in $($OCI network nat-gateway list --compartment-id "$COMP" --vcn-id "$vcn" --all --query "data[].id" 2>/dev/null | jq -r '.[]'); do
      echo "Deleting nat gateway $nat"
      retry_delete "nat gateway $nat" $OCI network nat-gateway delete --nat-gateway-id "$nat" --force
    done
    for sgw in $($OCI network service-gateway list --compartment-id "$COMP" --vcn-id "$vcn" --all --query "data[].id" 2>/dev/null | jq -r '.[]'); do
      echo "Deleting service gateway $sgw"
      retry_delete "service gateway $sgw" $OCI network service-gateway delete --service-gateway-id "$sgw" --force
    done
    echo "Deleting VCN $vcn"
    retry_delete "VCN $vcn" $OCI network vcn delete --vcn-id "$vcn" --force
done

echo "--- Tenancy-level IAM (filtered by prefix $PREFIX) ---"
for policy in $($OCI iam policy list --compartment-id "$TENANCY" --all --query "data[?starts_with(name, '${PREFIX}-')].id" 2>/dev/null | jq -r '.[]'); do
    echo "Deleting policy $policy"
    $OCI iam policy delete --policy-id "$policy" --force
done
for dg in $($OCI iam dynamic-group list --compartment-id "$TENANCY" --all --query "data[?starts_with(name, '${PREFIX}-')].id" 2>/dev/null | jq -r '.[]'); do
    echo "Deleting dynamic group $dg"
    $OCI iam dynamic-group delete --dynamic-group-id "$dg" --force
done

echo "--- Customer secret keys (own user, filtered by prefix $PREFIX) ---"
for key in $($OCI iam customer-secret-key list --user-id "$USER_OCID" \
  --query "data[?starts_with(\"display-name\", 'entigo-infralib-${PREFIX}')].id" 2>/dev/null | jq -r '.[]'); do
    echo "Deleting customer secret key $key"
    $OCI iam customer-secret-key delete --user-id "$USER_OCID" --customer-secret-key-id "$key" --force
done

echo "Done."
