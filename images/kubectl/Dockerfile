ARG TOOLS_TAG=latest
FROM entigolabs/entigo-infralib-tools:${TOOLS_TAG} AS tools
FROM alpine:3.22
RUN apk --no-cache add \
    ca-certificates \
    yq \
    jq \
    git \
    bash \
    sudo \
    curl \
    && rm -rf /var/cache/apk/* \
    && echo "entigo ALL=(ALL) NOPASSWD: /usr/sbin/update-ca-certificates" >> /etc/sudoers.d/entigo \
    && addgroup -g 1000 entigo && adduser -u 1000 -G entigo -D entigo && chown entigo /usr/local/share/ca-certificates
COPY --from=tools /entigo-bin/kubectl /usr/bin/
USER entigo
WORKDIR /app
