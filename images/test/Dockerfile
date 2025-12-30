ARG TOOLS_TAG=latest

FROM entigolabs/entigo-infralib-google:${TOOLS_TAG}
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ENV LOCAL_MODE=true ENTIGO_INFRALIB_DESTROY=false ENTIGO_INFRALIB_TEST_TIMEOUT=30m ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS=false ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS=false
ENV GOPATH="/go"
ENV PATH="/opt/venv/bin:${GOPATH}/bin:/usr/local/go/bin:${PATH}"
USER root
#Install AWS SDK
RUN apk --no-cache add \
    ca-certificates \
    python3 \
    py3-pip \
    yq \
    git \
    && rm -rf /var/cache/apk/* \
    && python3 -m venv /opt/venv \
    && /opt/venv/bin/pip install --no-cache-dir awscli \
    && rm -rf \
       /root/.cache/pip \
       /opt/venv/bin/pip* \
       /opt/venv/lib/python3.12/site-packages/awscli/examples \
       /opt/venv/lib/python3.12/site-packages/pip \
       /opt/venv/lib/python3.12/site-packages/setuptools \
       /opt/venv/lib/python3.12/site-packages/wheel \
       /opt/venv/lib/python3.12/site-packages/docutils \
       /opt/venv/lib/python3.12/site-packages/*/tests \
       /opt/venv/lib/python3.12/site-packages/*/test \
       /opt/venv/lib/python3.12/site-packages/botocore/docs \
       /opt/venv/lib/python3.12/site-packages/pip-* \
    && find /opt/venv/lib/python3.12/site-packages/botocore/data -name examples-1.json -delete \
    && find /opt/venv -type f -name "*.pyc" -delete \
    && find /opt/venv -type f -name "*.pyo" -delete \
    && apk del py3-pip
#Install golang
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      wget  --no-verbose https://go.dev/dl/go1.23.5.linux-arm64.tar.gz && tar -C /usr/local -xzf go1.23.5.linux-arm64.tar.gz && rm go1.23.5.linux-arm64.tar.gz; \
    elif [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
      wget  --no-verbose https://go.dev/dl/go1.23.5.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz && rm go1.23.5.linux-amd64.tar.gz; \
    fi
#SEE https://entigo.atlassian.net/browse/RD-76
ENV TF_PLUGIN_CACHE_DIR=/plugin-cache TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE="true"
COPY --chown=cloudsdk:cloudsdk cache /cache
COPY --chown=cloudsdk:cloudsdk test /app/test
COPY --chown=root:root common /common

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM" | tee -a /build_arch_testing && mkdir /plugin-cache /plan-json /go && chown -R cloudsdk:cloudsdk /plugin-cache /plan-json /cache /go

USER cloudsdk
RUN cd /cache && terraform init && ls -l / && ls -l /cache/ && rm -rf /cache/*
RUN export GOMAXPROCS=2 && cd /common && go mod download -x && cd /app && go mod init github.com/entigolabs/entigo-infralib && go mod edit -require github.com/entigolabs/entigo-infralib-common@v0.0.0 -replace github.com/entigolabs/entigo-infralib-common=/common && go mod tidy && cd test && go test -timeout 5m

