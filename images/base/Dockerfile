FROM debian:bookworm-slim
ARG TARGETPLATFORM
ARG BUILDPLATFORM
WORKDIR /app
ENV KUBECTL_VERSION=v1.29.8 TERRAFORM_VERSION=1.7.5-1 HELM_VERSION=v3.15.4 AWSCLI_VERSION=latest ARGOCD_CLI=v2.11.3 CLOUD_SDK_VERSION=499.0.0
ENTRYPOINT ["/usr/bin/entrypoint.sh"]  

ENV PATH=/google-cloud-sdk/bin:$PATH CLOUDSDK_PYTHON=/usr/bin/python3

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM" | tee -a /build_arch_base && apt-get update -y && apt-get -y install wget curl gpg && \
    wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com bullseye main" | tee /etc/apt/sources.list.d/hashicorp.list && \
    apt-get update -y && apt-get -y install unzip terraform=${TERRAFORM_VERSION} yq && rm -rf /var/cache/apt/archives /var/lib/apt/lists/* && \
    if [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      curl -Lo /usr/bin/argocd https://github.com/argoproj/argo-cd/releases/download/${ARGOCD_CLI}/argocd-linux-arm64 && chmod +x /usr/bin/argocd && \
      curl -s -o /usr/bin/kubectl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/arm64/kubectl" && chmod +x /usr/bin/kubectl && \
      curl -s "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip" && unzip awscliv2.zip && ./aws/install --bin-dir /usr/bin && rm -rf awscliv2.zip aws /usr/local/aws-cli/v2/current/dist/aws_completer /usr/local/aws-cli/v2/current/dist/awscli/data/ac.index /usr/local/aws-cli/v2/current/dist/awscli/examples glibc-*.apk && find /usr/local/aws-cli/v2/current/dist/awscli/botocore/data -name examples-1.json -delete && \
      curl -s "https://get.helm.sh/helm-${HELM_VERSION}-linux-arm64.tar.gz" -o "helm.tar.gz" && tar xzf helm.tar.gz && mv linux-arm64/helm /usr/bin/helm && rm -rf helm.tar.gz linux-* && chmod +x /usr/bin/helm && \
      cd / && curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${CLOUD_SDK_VERSION}-linux-arm.tar.gz && \
      tar xzf google-cloud-cli-${CLOUD_SDK_VERSION}-linux-arm.tar.gz && \
      rm google-cloud-cli-${CLOUD_SDK_VERSION}-linux-arm.tar.gz && \
      rm -rf /google-cloud-sdk/platform/bundledpythonunix && \
      gcloud config set core/disable_usage_reporting true && \
      gcloud config set component_manager/disable_update_check true && \
      gcloud config set metrics/environment github_docker_image && \
      gcloud components remove bq && \
      gcloud components install -q beta && \
      gcloud components install -q gke-gcloud-auth-plugin && \
      gcloud components update && \    
      rm -rf $(find google-cloud-sdk/ -regex ".*/__pycache__") && \
      rm -rf google-cloud-sdk/.install/.backup && \
      rm -rf google-cloud-sdk/bin/anthoscli && \
      git config --system credential.'https://source.developers.google.com'.helper gcloud.sh && \
      gcloud --version; \
    elif [ "$TARGETPLATFORM" = "linux/amd64" ]; then \
      curl -Lo /usr/bin/argocd https://github.com/argoproj/argo-cd/releases/download/${ARGOCD_CLI}/argocd-linux-amd64 && chmod +x /usr/bin/argocd && \
      curl -s -o /usr/bin/kubectl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && chmod +x /usr/bin/kubectl && \
      curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && unzip awscliv2.zip && ./aws/install --bin-dir /usr/bin && rm -rf awscliv2.zip aws /usr/local/aws-cli/v2/current/dist/aws_completer /usr/local/aws-cli/v2/current/dist/awscli/data/ac.index /usr/local/aws-cli/v2/current/dist/awscli/examples glibc-*.apk && find /usr/local/aws-cli/v2/current/dist/awscli/botocore/data -name examples-1.json -delete && \
      curl -s "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" -o "helm.tar.gz" && tar xzf helm.tar.gz && mv linux-amd64/helm /usr/bin/helm && rm -rf helm.tar.gz linux-* && chmod +x /usr/bin/helm && \
      cd / && curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
      tar xzf google-cloud-cli-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
      rm google-cloud-cli-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
      rm -rf /google-cloud-sdk/platform/bundledpythonunix && \
      gcloud config set core/disable_usage_reporting true && \
      gcloud config set component_manager/disable_update_check true && \
      gcloud config set metrics/environment github_docker_image && \
      gcloud components remove bq && \
      gcloud components install -q beta && \
      gcloud components install -q gke-gcloud-auth-plugin && \
      gcloud components update && \    
      rm -rf $(find google-cloud-sdk/ -regex ".*/__pycache__") && \
      rm -rf google-cloud-sdk/.install/.backup && \
      rm -rf google-cloud-sdk/bin/anthoscli && \
      git config --system credential.'https://source.developers.google.com'.helper gcloud.sh && \
      gcloud --version; \
    fi

COPY --from=entigolabs/entigo-infralib-agent /usr/bin/ei-agent /usr/bin/ei-agent

#SEE https://entigo.atlassian.net/browse/RD-76
COPY *.sh /usr/bin/


