## Oppinionated module for ECR container image registry proxies ##

Creates [ECR pull through cache rules](https://docs.aws.amazon.com/AmazonECR/latest/userguide/pull-through-cache.html) for the public container registries the infralib modules pull from. Images are pulled once from the upstream registry, stored in the account's private ECR and served from there on every following pull. This removes Docker Hub rate limits, keeps pulls inside the VPC through the ECR and S3 endpoints, and keeps working when the upstream registry is unreachable.

The other infralib modules detect this module and use the proxied registries automatically, for example `{{ .toptout.ecr-proxy.hub_registry | "docker.io" }}`. The `eks`, `eks-node-group` and `crossplane` modules attach the `policy` output so the nodes and Crossplane are allowed to pull through the caches and to create the cached repositories.

### Available pull through caches ###

The repository prefix is `<prefix>-<name>` where `<prefix>` is the module prefix (`<env>-<step>-<module>`, for example `biz-net-ecr-proxy`) cut to 24 characters.

| Name | Upstream registry | Repository prefix | Credentials | Output |
|------|-------------------|-------------------|-------------|--------|
| ecr  | public.ecr.aws    | `<prefix>-ecr`    | not needed, always created | `ecr_registry` |
| k8s  | registry.k8s.io   | `<prefix>-k8s`    | not needed, always created | `k8s_registry` |
| quay | quay.io           | `<prefix>-quay`   | not needed, always created | `quay_registry` |
| hub  | registry-1.docker.io (Docker Hub) | `<prefix>-hub` | required, created only when `hub_username` and `hub_token` are set | `hub_registry` |
| ghcr | ghcr.io (GitHub Container Registry) | `<prefix>-ghcr` | required, created only when `ghcr_username` and `ghcr_token` are set | `ghcr_registry` |
| gcr  | gcr.io (Google Container Registry) | `<prefix>-gcr` | required, created only when `gcr_username` and `gcr_token` are set | `gcr_registry` |

Each registry output is the full pull URL, `<account>.dkr.ecr.<region>.amazonaws.com/<prefix>-<name>`. Outputs of caches that are not created are `null`. Append the upstream image path to the output, for Docker Hub official images the `library/` namespace must be spelled out.

```
<account>.dkr.ecr.eu-north-1.amazonaws.com/biz-net-ecr-proxy-hub/library/nginx:1.29
<account>.dkr.ecr.eu-north-1.amazonaws.com/biz-net-ecr-proxy-hub/grafana/loki:3.5.0
<account>.dkr.ecr.eu-north-1.amazonaws.com/biz-net-ecr-proxy-quay/argoproj/argocd:v3.1.0
<account>.dkr.ecr.eu-north-1.amazonaws.com/biz-net-ecr-proxy-k8s/metrics-server/metrics-server:v0.8.0
<account>.dkr.ecr.eu-north-1.amazonaws.com/biz-net-ecr-proxy-ghcr/dexidp/dex:v2.44.0
```

Every cache gets an ECR repository creation template, so the repositories ECR creates on first pull have AES256 encryption, mutable tags and a lifecycle policy that expires untagged images after 7 days and tagged images 90 days after they were pushed into the cache. An image that is still in use is pulled again from upstream after it expires, it is not lost.

An additional cache for a private upstream ECR registry is created when `upstream_registry_url` is set, see below.

### Configuring the credentials ###

Docker Hub, GitHub Container Registry and Google Container Registry require upstream credentials. The module stores the username and token as a JSON `{"username": ..., "accessToken": ...}` Secrets Manager secret named `ecr-pullthroughcache/<prefix>-<name>` and points the cache rule at it. ECR requires that name prefix, do not rename the secret. The secret has a 7 day recovery window, so destroying and recreating the module with the same prefix inside 7 days fails until the old secret is force deleted.

Which token to use:

* __hub__ a Docker Hub username and a [personal access token](https://docs.docker.com/security/for-developers/access-tokens/) with read access. A free Docker Hub account is enough, the cache only needs to authenticate to avoid the anonymous rate limit.
* __ghcr__ a GitHub username and a classic personal access token with the `read:packages` scope.
* __gcr__ a Google service account key, see the [AWS documentation](https://docs.aws.amazon.com/AmazonECR/latest/userguide/pull-through-cache-creating-rule.html) for the username and token format ECR expects.

The tokens must not be written into the config file. Store them as agent custom parameters and reference them with the `output-custom` replacement tag. The agent stores custom parameters in AWS SSM Parameter Store. When the config contains a `kms` module that has already been applied once, the parameters are encrypted with that KMS key, so add the credentials after the first successful run of the `kms` module or they stay encrypted with the AWS managed key.

```
#Docker Hub
ei-agent add-custom --key=/ecr-proxy/hub/username --value=...
ei-agent add-custom --key=/ecr-proxy/hub/token --value=...
#Github
ei-agent add-custom --key=/ecr-proxy/ghcr/username --value=...
ei-agent add-custom --key=/ecr-proxy/ghcr/token --value=...
```

Then reference the parameters from the module inputs with the same keys.

```
    modules:
      - name: ecr-proxy
        source: aws/ecr-proxy
        inputs:
          hub_username: "{{ .output-custom./ecr-proxy/hub/username }}"
          hub_token: "{{ .output-custom./ecr-proxy/hub/token }}"
          ghcr_username: "{{ .output-custom./ecr-proxy/ghcr/username }}"
          ghcr_token: "{{ .output-custom./ecr-proxy/ghcr/token }}"
```

To rotate a token run `add-custom` again with `--overwrite=true` and run the agent, the new secret version is picked up by the cache rule. Without `--overwrite=true` an existing parameter is not replaced.

The parameter keys are free to choose, the keys above are the convention used in the infralib test environments. The same values can also be given as plain module inputs, which is only acceptable when the config file itself is a secret.

### Inputs ###

__hub_username__ and __hub_token__ Docker Hub credentials. Both must be set to create the `hub` cache, default empty.

__ghcr_username__ and __ghcr_token__ GitHub Container Registry credentials. Both must be set to create the `ghcr` cache, default empty.

__gcr_username__ and __gcr_token__ Google Container Registry credentials. Both must be set to create the `gcr` cache, default empty.

__upstream_registry_url__ Private ECR registry URL (`<account>.dkr.ecr.<region>.amazonaws.com`) to proxy with the repository prefix `ROOT`. Default empty. When set, the module also creates the `ECRPTCRole` IAM role that ECR assumes to pull from the upstream account, the upstream account must grant that role access to its repositories. Every repository in the upstream registry is reachable as `<account>.dkr.ecr.<region>.amazonaws.com/ROOT/<repository>`.

__upstream_registry_lifecycle_policy__ Lifecycle policy JSON applied to the repositories the `ROOT` cache creates. Defaults to the same 7 day untagged and 90 day tagged expiry as the public caches.

### Outputs ###

__hub_registry__, __ghcr_registry__, __gcr_registry__, __k8s_registry__, __ecr_registry__, __quay_registry__ Registry URL of each cache, `null` for caches that were not created.

__policy__ ARN of an IAM policy that allows pulling through the caches and creating the `<prefix>-*` repositories. Attach it to every role that pulls images, the `eks`, `eks-node-group` and `crossplane` modules do this on their own.

### Example code ###

Without credentials only the ecr, k8s and quay caches are created.

```
    modules:
      - name: ecr-proxy
        source: aws/ecr-proxy
```

With Docker Hub and GitHub Container Registry credentials from custom parameters.

```
    modules:
      - name: ecr-proxy
        source: aws/ecr-proxy
        inputs:
          hub_username: "{{ .output-custom./ecr-proxy/hub/username }}"
          hub_token: "{{ .output-custom./ecr-proxy/hub/token }}"
          ghcr_username: "{{ .output-custom./ecr-proxy/ghcr/username }}"
          ghcr_token: "{{ .output-custom./ecr-proxy/ghcr/token }}"
```

Proxying a private ECR registry of another account in addition to the public registries.

```
    modules:
      - name: ecr-proxy
        source: aws/ecr-proxy
        inputs:
          upstream_registry_url: "123456789012.dkr.ecr.eu-north-1.amazonaws.com"
```
