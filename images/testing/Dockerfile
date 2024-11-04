ARG BASE_IMAGE=latest
FROM --platform=${BUILDPLATFORM:-linux/amd64} entigolabs/entigo-infralib-base:$BASE_IMAGE
ENV ENTIGO_INFRALIB_DESTROY=false ENTIGO_INFRALIB_TEST_TIMEOUT=30m ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS=false ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS=false

COPY cache /cache  
#SEE https://entigo.atlassian.net/browse/RD-76
ENV TF_PLUGIN_CACHE_DIR=/plugin-cache TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE="true"
#SEE https://entigo.atlassian.net/browse/RD-76
RUN cd /cache && mkdir /plugin-cache && terraform init && rm -rf /cache 

COPY *.sh /usr/bin/
RUN apt update && apt -y install golang-go


COPY test /app/test
RUN BUILD=1 /usr/bin/entrypoint.sh
