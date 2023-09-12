FROM golang:1.20
WORKDIR /app
ENV KUBECTL_VERSION=v1.27.3 TERRAFORM_VERSION=1.5.2-1 HELM_VERSION=v3.10.2 AWSCLI_VERSION=latest
ENTRYPOINT ["/usr/bin/entrypoint.sh"]  
RUN wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com bullseye main" | tee /etc/apt/sources.list.d/hashicorp.list && \
    apt-get update && apt-get -y install unzip terraform=${TERRAFORM_VERSION} && \
    curl -s -o /usr/bin/kubectl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && chmod +x /usr/bin/kubectl && \
    curl -s "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && unzip awscliv2.zip && ./aws/install --bin-dir /usr/bin && \
    curl -s "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz" -o "helm.tar.gz" && \
    tar xzf helm.tar.gz && chmod +x linux-amd64/helm && mv linux-amd64/helm /usr/bin/helm
COPY cache-project /cache-project
#ENV TF_PLUGIN_CACHE_DIR=/plugin-cache  #SEE https://entigo.atlassian.net/browse/RD-76
#RUN cd /cache-project && mkdir /plugin-cache && terraform init  #SEE https://entigo.atlassian.net/browse/RD-76
COPY entrypoint.sh /usr/bin/entrypoint.sh 
