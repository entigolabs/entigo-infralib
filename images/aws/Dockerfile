ARG TOOLS_TAG=latest
FROM entigolabs/entigo-infralib-tools:${TOOLS_TAG} AS tools
FROM alpine:3.22
ENTRYPOINT ["/usr/bin/entrypoint.sh"]  
ENV PATH="/opt/venv/bin:$PATH"

#Install AWS SDK
RUN apk --no-cache add \
    ca-certificates \
    python3 \
    py3-pip \
    jq \
    yq \
    git \
    sudo \
    bash \
    && rm -rf /var/cache/apk/* \
    && python3 -m venv /opt/venv \
    && /opt/venv/bin/pip install --no-cache-dir awscli \
    && rm -rf \
       /root/.cache/pip \
       /opt/venv/bin/pip* \
       /opt/venv/lib/python3.12/site-packages/awscli/examples \
       /opt/venv/lib/python3.12/site-packages/pip \
       /opt/venv/lib/python3.12/site-packages/setuptools \
       /opt/venv/lib/python3.12/site-packages/wheel \
       /opt/venv/lib/python3.12/site-packages/*/tests \
       /opt/venv/lib/python3.12/site-packages/*/test \
       /opt/venv/lib/python3.12/site-packages/pip-* \
    && find /opt/venv/lib/python3.12/site-packages/botocore/data -name examples-1.json -delete \
    && find /opt/venv -type f -name "*.pyc" -delete \
    && find /opt/venv -type f -name "*.pyo" -delete \
    && apk del py3-pip \
    && echo "entigo ALL=(ALL) NOPASSWD: /usr/sbin/update-ca-certificates" >> /etc/sudoers.d/entigo \
    && addgroup -g 1000 entigo && adduser -u 1000 -G entigo -D entigo && chown entigo /usr/local/share/ca-certificates

# Copy stripped binaries from tools image
COPY --from=tools /entigo-bin/* /usr/bin/
COPY --from=tools /entigo-scripts/* /usr/bin/

USER entigo
WORKDIR /app
