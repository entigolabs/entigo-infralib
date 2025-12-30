ARG TOOLS_TAG=latest
FROM entigolabs/entigo-infralib-tools:${TOOLS_TAG} AS tools
FROM alpine:3.23
COPY --from=tools /entigo-bin/kubectl /usr/bin/
RUN addgroup -g 1000 entigo && adduser -u 1000 -G entigo -D entigo
USER entigo
