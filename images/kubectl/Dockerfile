ARG BASE_TAG=latest
FROM entigolabs/entigo-infralib-tools:${BASE_TAG} AS tools
FROM alpine:3.24
# The entrypoint of a pipeline step, as in the cloud images. The Jobs of the
# modules that use this image for kubectl set their own command and never
# reach it; for a shell, run the image with --entrypoint sh.
ENTRYPOINT ["/usr/bin/entrypoint.sh"]
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
# Everything a Kubernetes (ArgoCD) pipeline step needs on any cloud, without a
# cloud CLI: the step scripts, the agent that runs them, helm for the ArgoCD
# bootstrap, and shackle for the cluster, the registry and the bucket. No
# OpenTofu: an OpenTofu step runs in the cloud's own image.
COPY --from=tools /entigo-bin/kubectl /usr/bin/
COPY --from=tools /entigo-bin/argocd /usr/bin/
COPY --from=tools /entigo-bin/helm /usr/bin/
COPY --from=tools /entigo-bin/shackle /usr/bin/
COPY --from=tools /entigo-bin/ei-agent /usr/bin/
COPY --from=tools /entigo-scripts/* /usr/bin/
USER 1000
WORKDIR /app
