ARG AGENT_VERSION=v1.12.0-alpha3
FROM entigolabs/entigo-infralib-agent:${AGENT_VERSION} AS agent

FROM alpine:3.22
ARG TARGETARCH
ENV KUBECTL_VERSION=v1.35.5 TERRAFORM_VERSION=1.7.5 HELM_VERSION=v3.21.0 ARGOCD_CLI=v3.4.3 OPENTOFU_VERSION=1.12.1

COPY --from=agent /usr/bin/ei-agent /entigo-bin/ei-agent

RUN apk --no-cache add \
    unzip \
    binutils
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

RUN wget -O "opentofu.zip" "https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_${TARGETARCH}.zip" \
    && unzip opentofu.zip \
    && mv tofu /entigo-bin/ \
    && rm opentofu.zip \
    && strip /entigo-bin/tofu

COPY *.sh /entigo-scripts/

RUN addgroup -g 1000 entigo && adduser -u 1000 -G entigo -D entigo
USER 1000
