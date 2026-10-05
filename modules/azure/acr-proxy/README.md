## Azure Container Registry pull-through cache ##

Creates an ACR (`<prefix alphanumerics><agent uniqueSuffix>`, globally unique) with cache rules for Docker Hub, GitHub,
Google (gcr.io), AWS public ECR, Quay, Microsoft and Kubernetes registries. Images are pulled as
`<acr>.azurecr.io/<source registry>/<repository>`, see the `*_registry` outputs. AWS public ECR and Microsoft
(both only allow unauthenticated pulls in ACR) and Kubernetes registries are always anonymous.

A daily ACR task (`purge_*`, like the aws/ecr-proxy lifecycle policy) deletes untagged images after 7 days and all
tags after 90 days; they are pulled again from upstream on the next use. `acr purge` is a preview feature.

Like aws/ecr-proxy, all registries are anonymous by default. For credentials, store them as agent custom parameters
(`/acr-proxy/hub/username` becomes the Key Vault secret `acr-proxy-hub-username` in the agent's vault, `key_vault_id`)
and set the `*_secret` inputs to the secret names; each ACR credential set gets Key Vault Secrets User on its two
secrets. Docker Hub requires credentials: without them there is no hub cache and `hub_registry` is null. A secret name
that doesn't exist in the vault gives an unhealthy credential set.
```
ei-agent add-custom -k /acr-proxy/hub/username -v <username>
ei-agent add-custom -k /acr-proxy/hub/token -v <token>
```

**Removing credentials from a registry breaks its cache until the cached repositories are deleted.** Adding
credentials updates the cache rule in place. Removing them replaces the rule, and Azure rejects the new rule while
repositories cached by the old one exist (`TargetRepositoryPrefixMatchesExistingRegistries`), so that registry is not
proxied until they are deleted (images are pulled again from upstream on the next use). For Docker Hub the hub rule is
removed altogether, `hub_registry` becomes null and consumers fall back to docker.io. Before the apply, delete the
cached repositories of that registry (`docker.io`, `ghcr.io`, `gcr.io`, `public.ecr.aws` or `quay.io`):
```
ACR=<acr name>; SRC=ghcr.io
az acr repository list -n $ACR -o tsv | grep "^$SRC/" | xargs -r -I{} az acr repository delete -n $ACR --repository {} --yes
```

All tiers have cache rules; Basic has lower storage and request limits. `sku = "Premium"` is needed for
`private_endpoint_enabled` (pulls from the VNet use private IPs, the endpoint is in the vpc private subnet next to the
AKS nodes), customer managed keys (`encryption`, only at creation) and untagged manifest retention. The AKS bootstrap
artifact cache rule (`aks_bootstrap_cache_rule`) works on every tier, but aks `bootstrap_cache_enabled` needs a
Premium registry with the private endpoint.

The private endpoint gets its own `privatelink.azurecr.io` zone linked to `private_endpoint_vnet_id`. Only one zone
with that name can be linked to a VNet, so with a central zone (hub and spoke) or a second registry on the same VNet
set `private_dns_zone_id` to the existing, already linked zone.

Zone redundancy is not an input: Azure makes registries zone redundant in every region with availability zones (the
`zoneRedundancy` property still shows Disabled). The registry has no managed identity unless `encryption` is set.

### Example code ###
```
    modules:
      - name: acr-proxy
        source: azure/acr-proxy
        inputs:
          sku: "Premium"
          private_endpoint_enabled: true
          hub_username_secret: "acr-proxy-hub-username"
          hub_access_token_secret: "acr-proxy-hub-token"
```
