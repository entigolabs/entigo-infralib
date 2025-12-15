FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine
WORKDIR /app
ENTRYPOINT ["/usr/bin/entrypoint.sh"]

# Install additional tools
RUN apk --no-cache add \
    yq \
    git \
    && rm -rf /var/cache/apk/*

# Copy stripped binaries from tools image
COPY --from=entigolabs/entigo-infralib-base-tools:latest /entigo-bin/* /usr/bin/
COPY --from=entigolabs/entigo-infralib-base-tools:latest /entigo-scripts/* /usr/bin/
