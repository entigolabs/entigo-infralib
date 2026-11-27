ARG BASE_TAG=latest
FROM entigolabs/entigo-infralib-tools:${BASE_TAG} AS tools
FROM alpine:3.24
ENTRYPOINT ["/usr/bin/entrypoint.sh"]
# No cloud CLI: the step scripts reach the bucket, the registry and the cluster
# with shackle, and OpenTofu and its providers talk to the cloud themselves.
# That leaves this image the same for every cloud but Azure; it stays one per
# cloud because the agent picks a step's image by cloud and the test images
# are built on it.
RUN apk --no-cache add \
    ca-certificates \
    jq \
    yq \
    git \
    sudo \
    bash \
    && rm -rf /var/cache/apk/* \
    && echo "entigo ALL=(ALL) NOPASSWD: /usr/sbin/update-ca-certificates" >> /etc/sudoers.d/entigo \
    && addgroup -g 1000 entigo && adduser -u 1000 -G entigo -D entigo && chown entigo /usr/local/share/ca-certificates
# Copy stripped binaries from tools image
COPY --from=tools /entigo-bin/* /usr/bin/
COPY --from=tools /entigo-scripts/* /usr/bin/
USER 1000
WORKDIR /app
