ARG BASE_TAG=latest
FROM entigolabs/entigo-infralib-tools:${BASE_TAG} AS tools
FROM mcr.microsoft.com/azure-cli:azurelinux3.0
ARG TARGETARCH
ENTRYPOINT ["/usr/bin/entrypoint.sh"]
# The Azure CLI is still here for one thing: logging in as the managed
# identity of a job for OpenTofu (functions-azure.sh). The bucket, the
# registry and the cluster are reached with shackle, so kubelogin and azcopy
# are gone.
ENV YQ_VERSION=v4.53.6 \
    AZURE_CORE_COLLECT_TELEMETRY=false AZURE_CORE_ONLY_SHOW_ERRORS=true

RUN tdnf install -y \
    ca-certificates \
    jq \
    git \
    sudo \
    bash \
    unzip \
    tar \
    gawk \
    diffutils \
    shadow-utils \
    && tdnf clean all \
    && curl -fsSL -o /usr/bin/yq "https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_${TARGETARCH}" \
    && chmod +x /usr/bin/yq \
    # Azure Linux uses update-ca-trust, keep the update-ca-certificates interface used by functions-common.sh
    && mkdir -p /usr/local/share/ca-certificates \
    && printf '#!/bin/bash\ncp /usr/local/share/ca-certificates/*.crt /etc/pki/ca-trust/source/anchors/\nupdate-ca-trust\n' > /usr/sbin/update-ca-certificates \
    && chmod 755 /usr/sbin/update-ca-certificates \
    && echo "entigo ALL=(ALL) NOPASSWD: /usr/sbin/update-ca-certificates" >> /etc/sudoers.d/entigo \
    && groupadd -g 1000 entigo && useradd -u 1000 -g entigo -m entigo && chown entigo /usr/local/share/ca-certificates

# Copy stripped binaries from tools image
COPY --from=tools /entigo-bin/* /usr/bin/
COPY --from=tools /entigo-scripts/* /usr/bin/

USER 1000
WORKDIR /app
