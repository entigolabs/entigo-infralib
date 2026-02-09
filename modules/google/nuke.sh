#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

set -e

if [ "$GITHUB_ACTION" != "" ]; then

  docker run --rm \
    -e GOOGLE_APPLICATION_CREDENTIALS_JSON \
    -v "$SCRIPTPATH/google-nuke-config.yaml:/google-nuke-config.yaml:ro" \
    taivox/gcp-nuke:v0.0.4 \
    run --config /google-nuke-config.yaml --project-id entigo-infralib2 --no-prompt --prompt-delay 3 --quiet --no-dry-run --log-level trace

else

  docker run --rm \
    -v ~/.config/gcloud:/home/gcp-nuke/.config/gcloud:ro \
    -v "$(pwd)/google-nuke-config.yaml:/google-nuke-config.yaml:ro" \
    taivox/gcp-nuke:v0.0.4 \
    run --config /google-nuke-config.yaml --project-id entigo-infralib2 --no-prompt --prompt-delay 3 --quiet --no-dry-run --log-level trace

fi

