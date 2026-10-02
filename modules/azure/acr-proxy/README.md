## Azure Container Registry pull-through cache ##

Creates an ACR (`<prefix alphanumerics><4 char suffix>`, globally unique) with cache rules for Docker Hub, GitHub,
Google (gcr.io), AWS public ECR, Quay, Microsoft and Kubernetes registries. Images are pulled as
`<acr>.azurecr.io/<source registry>/<repository>`, see the `*_registry` outputs. AWS public ECR, Microsoft and
Kubernetes registries are always anonymous.

A daily ACR task (`purge_*`, like the aws/ecr-proxy lifecycle policy) deletes untagged images after 7 days and all
tags after 90 days; they are pulled again from upstream on the next use. `acr purge` is a preview feature.

Like aws/ecr-proxy, all registries are anonymous by default. For credentials, store them as agent custom parameters
(`/acr-proxy/hub/username` becomes the Key Vault secret `acr-proxy-hub-username` in the agent's vault, `key_vault_id`)
and set the `*_secret` inputs to the secret names; each ACR credential set gets Key Vault Secrets User on its two
secrets. Docker Hub requires credentials: without them there is no hub cache and `hub_registry` is null. A secret name
that doesn't exist in the vault gives an unhealthy credential set. Adding credentials to a registry updates its cache
rule in place, removing them replaces the rule, which Azure rejects while cached repositories exist under its path
(`TargetRepositoryPrefixMatchesExistingRegistries`): delete them first with `az acr repository delete`.
```
ei-agent add-custom -k /acr-proxy/hub/username -v <username>
ei-agent add-custom -k /acr-proxy/hub/token -v <token>
```

`sku = "Premium"` is needed for `private_endpoint_enabled` (pulls from the VNet use private IPs, the endpoint is in
the vpc private subnet next to the AKS nodes), the AKS bootstrap artifact cache rule (`aks_bootstrap_cache_rule`),
customer managed keys (`encryption`, only at creation) and untagged manifest retention.

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
