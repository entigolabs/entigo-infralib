#!/bin/bash

for var in "${!GIT_AUTH_SOURCE_@}"; do

	printf 'Configure git credentials for %s=%s\n' "$var" "${!var}"
        SOURCE="$(echo ${!var} | grep -oP '(?<=://)[^/]+')"
        PASSWORD="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_PASSWORD/g')"
        USERNAME="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_USERNAME/g')"
        git config --global url."https://${!USERNAME}:${!PASSWORD}@${SOURCE}".insteadOf https://${SOURCE}
        
done
