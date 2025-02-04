ARG BASE_IMAGE=latest
FROM entigolabs/entigo-infralib-base:$BASE_IMAGE
ARG TARGETPLATFORM
ARG BUILDPLATFORM

ENV ENTIGO_INFRALIB_DESTROY=false ENTIGO_INFRALIB_TEST_TIMEOUT=30m ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS=false ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS=false

COPY cache /cache  
#SEE https://entigo.atlassian.net/browse/RD-76
ENV TF_PLUGIN_CACHE_DIR=/plugin-cache TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE="true"
#SEE https://entigo.atlassian.net/browse/RD-76
RUN  echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM" | tee -a /build_arch_testing && cd /cache && mkdir /plugin-cache && terraform init && rm -rf /cache 

#RUN apt update && apt -y install golang-go
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      wget  --no-verbose https://go.dev/dl/go1.23.5.linux-arm64.tar.gz && tar -C /usr/local -xzf go1.23.5.linux-arm64.tar.gz && rm go1.23.5.linux-arm64.tar.gz; \
    elif [ "$TARGETPLATFORM" = "linux/amd64" ]; then \  
      wget  --no-verbose https://go.dev/dl/go1.23.5.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz && rm go1.23.5.linux-amd64.tar.gz; \
    fi
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

COPY test /app/test
COPY common /common
RUN export GOMAXPROCS=2 && cd /common && go mod download -x && cd /app && go mod init github.com/entigolabs/entigo-infralib && go mod edit -require github.com/entigolabs/entigo-infralib-common@v0.0.0 -replace github.com/entigolabs/entigo-infralib-common=/common && go mod tidy && cd test && go test -timeout 5m
COPY *.sh /usr/bin/
