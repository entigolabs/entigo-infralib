FROM alpine:3.22
ARG TARGETARCH
ENV KUBECTL_VERSION=v1.33.7 TERRAFORM_VERSION=1.7.5 HELM_VERSION=v3.19.4 ARGOCD_CLI=v3.1.10
RUN apk --no-cache add \
    unzip \
    binutils \
    && mkdir /entigo-bin
RUN wget -O /entigo-bin/argocd "https://github.com/argoproj/argo-cd/releases/download/${ARGOCD_CLI}/argocd-linux-${TARGETARCH}" \
    && chmod +x /entigo-bin/argocd \
    && strip /entigo-bin/argocd

RUN wget -O /entigo-bin/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl" \
    && chmod +x /entigo-bin/kubectl \
    && strip /entigo-bin/kubectl

RUN wget -O "helm.tar.gz" "https://get.helm.sh/helm-${HELM_VERSION}-linux-${TARGETARCH}.tar.gz" \
    && tar xzf helm.tar.gz \
    && mv linux-${TARGETARCH}/helm /entigo-bin/helm \
    && rm -rf helm.tar.gz linux-* \
    && chmod +x /entigo-bin/helm \
    && strip /entigo-bin/helm

RUN wget -O "terraform.zip" "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip" \
    && unzip terraform.zip \
    && mv terraform /entigo-bin/ \
    && rm terraform.zip \
    && strip /entigo-bin/terraform

#COPY --from=entigolabs/entigo-infralib-agent:latest /usr/bin/ei-agent /entigo-bin/ei-agent
#RUN strip /entigo-bin/ei-agent

COPY *.sh /entigo-scripts/

RUN addgroup -g 1000 entigo && adduser -u 1000 -G entigo -D entigo
USER entigo
