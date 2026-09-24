#!/bin/bash
# Removes every Entigo Infralib resource from the Oracle Cloud test compartment.
#
# The compartment must also be listed in oci-nuke-config.yml; --compartment-id alone is
# refused. A compartment still holding certificates ends the run with its CA failed and a
# non-zero exit, and only a run a day later can finish it - retrying now will not.
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd "$SCRIPTPATH" || exit 1

OCI_NUKE_IMAGE="${OCI_NUKE_IMAGE:-ghcr.io/entigolabs/oci-nuke:0.1.8}"
ORACLE_COMPARTMENT_ID="${ORACLE_COMPARTMENT_ID:-ocid1.compartment.oc1..aaaaaaaa4s6svm4opv5vovkdccgs72xlkmfab25tmblrszb6weyk6qpt255q}"

if [ "$PREFIX" == "" ]
then
  echo "ERROR: PREFIX must be set. Every tenancy-scoped lister - dynamic groups, policies,"
  echo "       users, groups, customer secret keys - returns nothing at all when the prefix"
  echo "       is empty, so the run would sweep the compartment, silently leave this"
  echo "       deployment's tenancy-wide IAM standing, and still exit 0 as if it were clean."
  exit 1
fi

if [ "$OCI_REGION" == "" ]
then
  echo "Defaulting OCI_REGION to eu-frankfurt-1"
  export OCI_REGION="eu-frankfurt-1"
fi

if [ "$OCI_CONFIG_FILE" == "" ]
then
  echo "Defaulting OCI_CONFIG_FILE to $(echo ~)/.oci/config"
  export OCI_CONFIG_FILE="$(echo ~)/.oci/config"
fi
OCI_CONFIG_DIR="$(dirname "$OCI_CONFIG_FILE")"

NUKE_OPTS=""
if [ "$DRY_RUN" != "true" ]
then
  NUKE_OPTS="--no-dry-run"
fi

DOCKER_OPTS=""
PROMPT_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
  # Interactive runs keep the prompt, which asks for the compartment ID to be typed back.
  DOCKER_OPTS="-it"
else
  PROMPT_OPTS="--no-prompt"
fi

echo "Nuking Oracle Cloud test resources"
echo "  compartment: $ORACLE_COMPARTMENT_ID"
echo "  region:      $OCI_REGION"
echo "  prefix:      $PREFIX"
echo "  image:       $OCI_NUKE_IMAGE"

# Mounted twice on purpose: at the image user's ~/.oci where the SDK looks, and at its own
# host path, because key_file= and security_token_file= in the config are absolute.
# The image's own user is uid 1000, and the mounted config is 0600, so the container can
# only read it when it runs as the uid that wrote it.
docker run --rm $DOCKER_OPTS --user "$(id -u):$(id -g)" \
	-v "$OCI_CONFIG_DIR":"/home/oci-nuke/.oci":ro \
	-v "$OCI_CONFIG_DIR":"$OCI_CONFIG_DIR":ro \
	-v "$SCRIPTPATH/oci-nuke-config.yml":"/home/oci-nuke/config.yml":ro \
	-e OCI_CONFIG_FILE="$OCI_CONFIG_FILE" \
	-e OCI_REGION="$OCI_REGION" \
	"$OCI_NUKE_IMAGE" run \
	--config /home/oci-nuke/config.yml \
	--compartment-id "$ORACLE_COMPARTMENT_ID" \
	--region "$OCI_REGION" \
	--prefix "$PREFIX" \
	$NUKE_OPTS $PROMPT_OPTS "$@"
