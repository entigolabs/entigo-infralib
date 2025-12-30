ARG TOOLS_TAG=latest
FROM entigolabs/entigo-infralib-tools:${TOOLS_TAG} AS tools
FROM alpine:latest
COPY --from=tools /entigo-bin/kubectl /usr/bin/
