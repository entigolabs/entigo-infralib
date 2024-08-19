FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.20
WORKDIR /app
ENV KUBECTL_VERSION=v1.29.8 TERRAFORM_VERSION=1.7.5-1 HELM_VERSION=v3.15.4 AWSCLI_VERSION=latest ARGOCD_CLI=v2.11.3
ENTRYPOINT ["/usr/bin/entrypoint.sh"]  
ARG TARGETPLATFORM=linux/amd64

RUN wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com bullseye main" | tee /etc/apt/sources.list.d/hashicorp.list && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee /etc/apt/sources.list.d/google-cloud-sdk.list && \ 
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg && \
    apt-get update -y && apt-get -y install unzip terraform=${TERRAFORM_VERSION} google-cloud-sdk google-cloud-sdk-gke-gcloud-auth-plugin yq && \
    if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      curl -Lo /usr/bin/argocd https://github.com/argoproj/argo-cd/releases/download/${ARGOCD_CLI}/argocd-linux-arm64 && chmod +x /usr/bin/argocd && \
      curl -s -o /usr/bin/kubectl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/arm64/kubectl" && chmod +x /usr/bin/kubectl && \
      curl -s "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip" && unzip awscliv2.zip && ./aws/install --bin-dir /usr/bin && \
      curl -s "https://get.helm.sh/helm-${HELM_VERSION}-linux-arm64.tar.gz" -o "helm.tar.gz" && tar xzf helm.tar.gz && mv linux-arm64/helm /usr/bin/helm && chmod +x /usr/bin/helm; \
    elif [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
      curl -Lo /usr/bin/argocd https://github.com/argoproj/argo-cd/releases/download/${ARGOCD_CLI}/argocd-linux-amd64 && chmod +x /usr/bin/argocd && \
      curl -s -o /usr/bin/kubectl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && chmod +x /usr/bin/kubectl && \
      curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && unzip awscliv2.zip && ./aws/install --bin-dir /usr/bin && \
      curl -s "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" -o "helm.tar.gz" && tar xzf helm.tar.gz && mv linux-amd64/helm /usr/bin/helm && chmod +x /usr/bin/helm; \
    fi
COPY cache /cache  
#SEE https://entigo.atlassian.net/browse/RD-76
ENV TF_PLUGIN_CACHE_DIR=/plugin-cache TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE="true"
#SEE https://entigo.atlassian.net/browse/RD-76
RUN cd /cache && mkdir /plugin-cache && terraform init && rm -rf /cache 
#SEE https://entigo.atlassian.net/browse/RD-76
COPY *.sh /usr/bin/


