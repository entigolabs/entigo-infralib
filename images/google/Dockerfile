ARG TOOLS_TAG=latest
FROM entigolabs/entigo-infralib-tools:${TOOLS_TAG} AS tools
FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine
ENTRYPOINT ["/usr/bin/entrypoint.sh"]

# Install additional tools
RUN apk --no-cache add \
    yq \
    jq \
    git \
    && rm -rf /var/cache/apk/* && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud config set metrics/environment github_docker_image && \
    gcloud components install -q beta && \
    gcloud components install -q gke-gcloud-auth-plugin && \
    gcloud components update && \
    git config --system credential.'https://source.developers.google.com'.helper gcloud.sh

# Copy stripped binaries from tools image
COPY --from=tools /entigo-bin/* /usr/bin/
COPY --from=tools /entigo-scripts/* /usr/bin/

USER cloudsdk
WORKDIR /app
