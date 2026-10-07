## Oppinionated module for Google Artifact Registry proxies ##

Creates [Artifact Registry remote repositories](https://cloud.google.com/artifact-registry/docs/repositories/remote-repo) for the public container registries the infralib modules pull from. Images are pulled once from the upstream registry, stored in the project's Artifact Registry and served from there on every following pull. This removes Docker Hub rate limits, keeps pulls inside Google's network through Private Google Access, and keeps working when the upstream registry is unreachable.

The other infralib modules detect this module and use the proxied registries automatically, for example `{{ .toptout.gar-proxy.hub_registry | "docker.io" }}`. GKE node service accounts and the `crossplane` module get the `roles/artifactregistry.reader` role from their own modules, no extra permission is needed to pull.

### Available proxies ###

All six remote repositories are always created. The repository id is `<prefix>-<name>` where `<prefix>` is the module prefix (`<env>-<step>-<module>`, for example `biz-net-gar-proxy`) cut to 50 characters.

| Name | Upstream registry | Repository id | Credentials | Output |
|------|-------------------|---------------|-------------|--------|
| hub  | registry-1.docker.io (Docker Hub) | `<prefix>-hub` | optional, `hub_username_secret` and `hub_access_token_secret` | `hub_registry` |
| ghcr | ghcr.io (GitHub Container Registry) | `<prefix>-ghcr` | optional, `ghcr_username_secret` and `ghcr_access_token_secret` | `ghcr_registry` |
| gcr  | gcr.io (Google Container Registry) | `<prefix>-gcr` | optional, `gcr_username_secret` and `gcr_access_token_secret` | `gcr_registry` |
| ecr  | public.ecr.aws    | `<prefix>-ecr`  | optional, `ecr_username_secret` and `ecr_access_token_secret` | `ecr_registry` |
| quay | quay.io           | `<prefix>-quay` | optional, `quay_username_secret` and `quay_access_token_secret` | `quay_registry` |
| k8s  | registry.k8s.io   | `<prefix>-k8s`  | not supported | `k8s_registry` |

Each registry output is the full pull URL, `<region>-docker.pkg.dev/<project>/<prefix>-<name>`, where the region is the provider region of the step. Append the upstream image path to the output, for Docker Hub official images the `library/` namespace must be spelled out.

```
europe-north1-docker.pkg.dev/my-project/biz-net-gar-proxy-hub/library/nginx:1.29
europe-north1-docker.pkg.dev/my-project/biz-net-gar-proxy-hub/grafana/loki:3.5.0
europe-north1-docker.pkg.dev/my-project/biz-net-gar-proxy-quay/argoproj/argocd:v3.1.0
europe-north1-docker.pkg.dev/my-project/biz-net-gar-proxy-k8s/metrics-server/metrics-server:v0.8.0
europe-north1-docker.pkg.dev/my-project/biz-net-gar-proxy-ghcr/dexidp/dex:v2.44.0
```

Every repository has cleanup policies that delete untagged images older than 7 days and all images older than 90 days. An image that is still in use is pulled again from upstream after it is deleted, it is not lost. Vulnerability scanning of the cached images is off by default, see `enablement_config`.

### Configuring the credentials ###

Anonymous pulls work for every upstream except that Docker Hub applies its anonymous rate limit per source IP, so Docker Hub credentials are strongly recommended. Credentials for the other upstreams are only needed for private images or higher rate limits.

Unlike the AWS `ecr-proxy` module this module does not take the credential values as inputs. It takes the **names of Secret Manager secrets** in the same project, one holding the username and one holding the access token. The module reads the username secret, passes the latest version of the access token secret to Artifact Registry and grants the Artifact Registry service agent `roles/secretmanager.secretAccessor` on it. A proxy gets credentials only when both of its secret inputs are set.

Which token to use:

* __hub__ a Docker Hub username and a [personal access token](https://docs.docker.com/security/for-developers/access-tokens/) with read access. A free Docker Hub account is enough.
* __ghcr__ a GitHub username and a classic personal access token with the `read:packages` scope.
* __quay__, __gcr__, __ecr__ a robot account, service account key or registry token of the upstream, see the [Artifact Registry documentation](https://cloud.google.com/artifact-registry/docs/repositories/remote-repo#docker) for the format each upstream expects.

Create the secrets with the agent `add-custom` command. On Google it creates a Secret Manager secret named exactly like the key, so the key must be a valid secret id, letters, digits, `-` and `_` only. The secrets are encrypted with the customer KMS key when the config contains a `kms` module that has already been applied once, so add the credentials after the first successful run of the `kms` module or they stay encrypted with the Google managed key.

```
#Docker Hub
ei-agent add-custom --key=gar-proxy-hub-username --value=...
ei-agent add-custom --key=gar-proxy-hub-access-token --value=...
#Github
ei-agent add-custom --key=gar-proxy-ghcr-username --value=...
ei-agent add-custom --key=gar-proxy-ghcr-access-token --value=...
```

Then give the secret names to the module. These are plain names, not `output-custom` tags, because the module reads the secrets itself and Artifact Registry needs the secret to exist for the token.

```
    modules:
      - name: gar-proxy
        source: google/gar-proxy
        inputs:
          hub_username_secret: "gar-proxy-hub-username"
          hub_access_token_secret: "gar-proxy-hub-access-token"
          ghcr_username_secret: "gar-proxy-ghcr-username"
          ghcr_access_token_secret: "gar-proxy-ghcr-access-token"
```

To rotate a token run `add-custom` again with the new value and run the agent, the repository is updated to the new secret version. Without `--overwrite=true` the command asks for confirmation before replacing an existing secret. The secret names are free to choose, the names above are the convention used in the infralib test environments. Secrets created by other means work as well, as long as the agent and the Artifact Registry service agent can read them.

### Inputs ###

__hub_username_secret__ and __hub_access_token_secret__ Secret Manager secret names with the Docker Hub credentials. Both must be set to use them, default empty.

__ghcr_username_secret__ and __ghcr_access_token_secret__ Secret Manager secret names with the GitHub Container Registry credentials. Both must be set to use them, default empty.

__gcr_username_secret__ and __gcr_access_token_secret__ Secret Manager secret names with the Google Container Registry credentials. Both must be set to use them, default empty.

__ecr_username_secret__ and __ecr_access_token_secret__ Secret Manager secret names with the Amazon ECR Public credentials. Both must be set to use them, default empty.

__quay_username_secret__ and __quay_access_token_secret__ Secret Manager secret names with the Quay credentials. Both must be set to use them, default empty.

__enablement_config__ Artifact Registry vulnerability scanning of the cached images, `DISABLED` or `INHERITED`. Defaults to `DISABLED`, scanning is billed per image.

### Outputs ###

__hub_registry__, __ghcr_registry__, __gcr_registry__, __ecr_registry__, __quay_registry__, __k8s_registry__ Registry URL of each proxy.

### Example code ###

Without credentials every proxy pulls anonymously.

```
    modules:
      - name: gar-proxy
        source: google/gar-proxy
```

With Docker Hub and GitHub Container Registry credentials from Secret Manager secrets.

```
    modules:
      - name: gar-proxy
        source: google/gar-proxy
        inputs:
          hub_username_secret: "gar-proxy-hub-username"
          hub_access_token_secret: "gar-proxy-hub-access-token"
          ghcr_username_secret: "gar-proxy-ghcr-username"
          ghcr_access_token_secret: "gar-proxy-ghcr-access-token"
```
