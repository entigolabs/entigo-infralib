#!/bin/bash
set -x
for var in "${!GIT_AUTH_SOURCE_@}"; do

	printf 'Configure git credentials for %s=%s\n' "$var" "${!var}"
        SOURCE="$(echo ${!var} | sed 's#git::https://##g' | sed 's/\.git.*$/.git/')"
        PASSWORD="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_PASSWORD/g')"
        USERNAME="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_USERNAME/g')"
        git config --global url."https://${!USERNAME}:${!PASSWORD}@${SOURCE}".insteadOf ${SOURCE}
        
done
