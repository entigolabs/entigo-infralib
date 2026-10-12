# Dockerfile for isolated Claude Code dependency updates
FROM node:20-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    jq \
    ca-certificates \
    openssh-client \
    && rm -rf /var/lib/apt/lists/*

# Pre-populate GitHub SSH host key
RUN mkdir -p /etc/ssh && ssh-keyscan github.com >> /etc/ssh/ssh_known_hosts 2>/dev/null

# Install Claude Code via npm
RUN npm install -g @anthropic-ai/claude-code

# Install helm
RUN curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install yq
RUN curl -fsSL -o /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 \
    && chmod +x /usr/local/bin/yq

# Install gh CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
    | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
    | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update && apt-get install -y gh && rm -rf /var/lib/apt/lists/*

# node:20-slim already has a 'node' user with UID 1000 — reuse it
RUN mkdir -p /home/node/.claude /home/node/.local/share/claude /home/node/.ssh /home/node/.config/gh \
    && echo '{}' > /home/node/.claude.json \
    && chown -R node:node /home/node

USER node

WORKDIR /workspace

ENTRYPOINT ["claude"]
