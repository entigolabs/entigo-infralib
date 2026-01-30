#!/bin/bash

set -e

if [ "$GITHUB_ACTION" != "" ]; then

  docker run --rm \
    -e GOOGLE_APPLICATION_CREDENTIALS_JSON \
    -v "$(pwd)/google-nuke-config.yaml:/google-nuke-config.yaml:ro" \
    ghcr.io/ekristen/gcp-nuke:v1.11.1 \
    run --config /google-nuke-config.yaml --project-id entigo-infralib2 --no-prompt --prompt-delay 3

  docker run --rm \
    -e GOOGLE_APPLICATION_CREDENTIALS_JSON \
    -v "$(pwd)/google-nuke-config.yaml:/google-nuke-config.yaml:ro" \
    ghcr.io/ekristen/gcp-nuke:v1.11.1 \
    run --config /google-nuke-config.yaml --project-id entigo-infralib2 --no-prompt --prompt-delay 3

else

  docker run --rm \
    -v ~/.config/gcloud:/home/gcp-nuke/.config/gcloud:ro \
    -v "$(pwd)/google-nuke-config.yaml:/google-nuke-config.yaml:ro" \
    ghcr.io/ekristen/gcp-nuke:v1.11.1 \
    run --config /google-nuke-config.yaml --project-id entigo-infralib2 --no-prompt --prompt-delay 3

  docker run --rm \
    -v ~/.config/gcloud:/home/gcp-nuke/.config/gcloud:ro \
    -v "$(pwd)/google-nuke-config.yaml:/google-nuke-config.yaml:ro" \
    ghcr.io/ekristen/gcp-nuke:v1.11.1 \
    run --config /google-nuke-config.yaml --project-id entigo-infralib2 --no-prompt --prompt-delay 3

fi

