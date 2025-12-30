ARG TOOLS_TAG=latest
FROM entigolabs/entigo-infralib-tools:${TOOLS_TAG} AS tools
FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine
WORKDIR /app
ENTRYPOINT ["/usr/bin/entrypoint.sh"]

# Install additional tools
RUN apk --no-cache add \
    yq \
    git \
    && rm -rf /var/cache/apk/*

# Copy stripped binaries from tools image
COPY --from=tools /entigo-bin/* /usr/bin/
COPY --from=tools /entigo-scripts/* /usr/bin/

RUN addgroup -g 1000 entigo && adduser -u 1000 -G entigo -D entigo
USER entigo
