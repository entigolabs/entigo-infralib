ARG TOOLS_TAG=latest
FROM entigolabs/entigo-infralib-tools:${TOOLS_TAG} AS tools
FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine
ENTRYPOINT ["/usr/bin/entrypoint.sh"]
ENV PATH="/opt/venv/bin:$PATH"

# Install additional tools
RUN apk --no-cache add \
    ca-certificates \
    python3 \
    py3-pip \
    jq \
    git \
    sudo \
    && rm -rf /var/cache/apk/* && \
    python3 -m venv /opt/venv && \
    /opt/venv/bin/pip install --no-cache-dir yq && \
    rm -rf \
       /root/.cache/pip \
       /opt/venv/bin/pip* \
       /opt/venv/lib/python3.12/site-packages/pip \
       /opt/venv/lib/python3.12/site-packages/setuptools \
       /opt/venv/lib/python3.12/site-packages/wheel \
       /opt/venv/lib/python3.12/site-packages/*/tests \
       /opt/venv/lib/python3.12/site-packages/*/test \
       /opt/venv/lib/python3.12/site-packages/pip-* && \
    find /opt/venv -type f -name "*.pyc" -delete && \
    find /opt/venv -type f -name "*.pyo" -delete && \
    apk del py3-pip && \
    sudo -u cloudsdk gcloud config set core/disable_usage_reporting true && \
    sudo -u cloudsdk gcloud config set component_manager/disable_update_check true && \
    sudo -u cloudsdk gcloud config set metrics/environment github_docker_image && \
    gcloud components install -q beta && \
    gcloud components install -q gke-gcloud-auth-plugin && \
    gcloud components update && \
    git config --system credential.'https://source.developers.google.com'.helper gcloud.sh && \
    echo "cloudsdk ALL=(ALL) NOPASSWD: /usr/sbin/update-ca-certificates" >> /etc/sudoers.d/cloudsdk && \
    chown cloudsdk /usr/local/share/ca-certificates

# Copy stripped binaries from tools image
COPY --from=tools /entigo-bin/* /usr/bin/
COPY --from=tools /entigo-scripts/* /usr/bin/

USER cloudsdk
WORKDIR /app
