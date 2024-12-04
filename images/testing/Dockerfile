ARG BASE_IMAGE=latest
FROM --platform=${BUILDPLATFORM:-linux/amd64} entigolabs/entigo-infralib-base:$BASE_IMAGE
ENV ENTIGO_INFRALIB_DESTROY=false ENTIGO_INFRALIB_TEST_TIMEOUT=30m ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS=false ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS=false

COPY cache /cache  
#SEE https://entigo.atlassian.net/browse/RD-76
ENV TF_PLUGIN_CACHE_DIR=/plugin-cache TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE="true"
#SEE https://entigo.atlassian.net/browse/RD-76
RUN cd /cache && mkdir /plugin-cache && terraform init && rm -rf /cache 

#RUN apt update && apt -y install golang-go
RUN wget  --no-verbose https://go.dev/dl/go1.23.4.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz && rm go1.23.4.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

COPY *.sh /usr/bin/
COPY test /app/test
COPY common /common
RUN /usr/bin/entrypoint.sh
