ARG BASE_IMAGE=latest
FROM --platform=${BUILDPLATFORM:-linux/amd64} entigolabs/entigo-infralib-base:$BASE_IMAGE
ENV ENTIGO_INFRALIB_DESTROY=false ENTIGO_INFRALIB_TEST_TIMEOUT=30m ENTIGO_INFRALIB_KUBECTL_EKS_CONTEXTS=false ENTIGO_INFRALIB_KUBECTL_GKE_CONTEXTS=false BASE_IMAGE=$BASE_IMAGE
COPY entrypoint.sh /usr/bin/entrypoint.sh
COPY test /app/test
RUN BUILD=1 /usr/bin/entrypoint.sh
