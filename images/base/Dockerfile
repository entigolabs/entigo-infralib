FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.20
WORKDIR /app
ENV KUBECTL_VERSION=v1.27.3 TERRAFORM_VERSION=1.5.2-1 HELM_VERSION=v3.10.2 AWSCLI_VERSION=latest
ENTRYPOINT ["/usr/bin/entrypoint.sh"]  
ARG TARGETPLATFORM

RUN wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com bullseye main" | tee /etc/apt/sources.list.d/hashicorp.list && \
    apt-get update && apt-get -y install unzip terraform=${TERRAFORM_VERSION} && \
    if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      curl -s -o /usr/bin/kubectl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/arm64/kubectl" && chmod +x /usr/bin/kubectl && \
      curl -s "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip" && unzip awscliv2.zip && ./aws/install --bin-dir /usr/bin && \
      curl -s "https://get.helm.sh/helm-${HELM_VERSION}-linux-arm64.tar.gz" -o "helm.tar.gz" && tar xzf helm.tar.gz && mv linux-arm64/helm /usr/bin/helm && chmod +x /usr/bin/helm; \
    elif [ "$TARGETPLATFORM" = "linux/amd64"]; then \
      curl -s -o /usr/bin/kubectl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && chmod +x /usr/bin/kubectl && \
      curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && unzip awscliv2.zip && ./aws/install --bin-dir /usr/bin && \
      curl -s "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" -o "helm.tar.gz" && tar xzf helm.tar.gz && mv linux-amd64/helm /usr/bin/helm && chmod +x /usr/bin/helm; \
    fi
#COPY cache /cache  #SEE https://entigo.atlassian.net/browse/RD-76
#ENV TF_PLUGIN_CACHE_DIR=/plugin-cache  #SEE https://entigo.atlassian.net/browse/RD-76
#RUN cd /cache && mkdir /plugin-cache && terraform init && rm -rf /cache #SEE https://entigo.atlassian.net/browse/RD-76
COPY entrypoint.sh /usr/bin/entrypoint.sh 
