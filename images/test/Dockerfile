ARG BASE_TAG=latest

# Placeholders for two directories of the repo that live outside this build
# context: the shared Go test library common/ and the module tree modules/.
# The CI build overrides them with --build-context common=<repo>/common and
# --build-context modules=<repo>/modules, and so must a local build; see
# README.md. They are only used to warm the Go module and build caches so that
# module tests download nothing at run time. At run time the real common/ is
# bind-mounted over /common by common/*/unit.sh.
FROM scratch AS common
FROM scratch AS modules

# Keep only the module test sources, so the final image does not carry the
# whole module tree.
FROM alpine:3.22 AS module-tests
COPY --from=modules . /src
RUN mkdir /warm && cd /src && find . -path '*/test/*_test.go' | while read -r f; do \
      mkdir -p "/warm/$(dirname "$f")" && cp "$f" "/warm/$f"; \
    done

FROM entigolabs/entigo-infralib-google:${BASE_TAG} AS image
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ENV LOCAL_MODE=true ENTIGO_INFRALIB_DESTROY=false ENTIGO_INFRALIB_TEST_TIMEOUT=30m ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS=false ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS=false
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:/usr/local/go/bin:${PATH}"
# Never download a different Go toolchain at run time. If common/go.mod or a
# dependency starts requiring a newer Go than the one installed below, fail
# loudly here and bump the version instead of fetching ~70MB on every test run.
ENV GOTOOLCHAIN=local
USER root
#Install AWS SDK
RUN apk --no-cache add \
    py3-pip \
    docker-cli \
    docker-credential-ecr-login \
    && rm -rf /var/cache/apk/* \
    && pip install --no-cache-dir --break-system-packages awscli \
    && rm -rf \
       /root/.cache/pip \
       /usr/lib/python3.*/site-packages/awscli/examples \
       /usr/lib/python3.*/site-packages/*/tests \
    && find /usr/lib/python3.* -type f -name examples-1.json -delete \
    && find /usr/lib/python3.* -type f -name "*.pyc" -delete \
    && apk del py3-pip
#Install golang. Must satisfy the go directive in common/go.mod, and must
# match what its pinned dependencies compile with: x/net v0.54.0 drops the
# http2 server under a go1.27 build constraint, which breaks grpc v1.68.0, so
# stay on 1.26.x until common/go.mod moves those on.
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      wget  --no-verbose https://go.dev/dl/go1.26.8.linux-arm64.tar.gz && tar -C /usr/local -xzf go1.26.8.linux-arm64.tar.gz && rm go1.26.8.linux-arm64.tar.gz; \
    elif [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
      wget  --no-verbose https://go.dev/dl/go1.26.8.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.26.8.linux-amd64.tar.gz && rm go1.26.8.linux-amd64.tar.gz; \
    fi
#SEE https://entigo.atlassian.net/browse/RD-76
ENV TF_PLUGIN_CACHE_DIR=/plugin-cache TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE="true"
COPY cache /cache
COPY test /app/test
COPY --from=common . /common
COPY --from=module-tests /warm /warm
COPY warm-go-cache.sh /usr/local/bin/warm-go-cache.sh

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM" | tee -a /build_arch_testing && mkdir /plugin-cache /plan-json /go

RUN cd /cache && tofu init -reconfigure && ls -l / && ls -l /cache/ && rm -rf /cache/*
RUN warm-go-cache.sh && rm -rf /warm
