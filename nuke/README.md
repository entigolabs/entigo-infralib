# Nuke

Every evening at 16:00 UTC the `Nuke` workflow deletes everything in the entigo-infralib test accounts that these files do not keep, one job per cloud, so that `Agent Stable` provisions the environments from scratch in the morning. This keeps costs down and proves a clean installation daily.

The files are the configurations of the nuke tools, in their own formats:

| File | Tool | Keeps, among other things |
|---|---|---|
| `aws.yaml` | [aws-nuke](https://github.com/ekristen/aws-nuke) | the CI and maintainer IAM users and their keys, the `infralib.entigo.io` zone, SSO roles, the saml-proxy secrets |
| `google.yaml` | [gcp-nuke](https://github.com/taivox/gcp-nuke) | the github service accounts, the `gcp.infralib.entigo.io` zone, the Artifact Registry repository |
| `oracle.yaml` | [oci-nuke](https://github.com/entigolabs/oci-nuke) | the agent's state keys, the bootstrap policies, the vaults, keys and CAs |
| `azure.yaml` | [azure-nuke](https://github.com/taivox/azure-nuke) | the `infralib-permanent-*` resource groups (parent DNS zones, AKS private DNS zone) with only the zones' apex NS/SOA records and no build identities' role assignments, resource groups outside `westeurope` and `swedencentral`, the soft-deleted Key Vaults (the agent recovers its vault with the registry credentials; after 7 days without a stable run, run `ei-agent add-custom` again) |

The workflow is [module-nuke.yaml](https://github.com/entigolabs/entigo-infralib-test/blob/main/.github/workflows/module-nuke.yaml) of entigo-infralib-test, which runs the tools as containers through its `scripts/nuke.sh`. Nuking a single cloud by hand: run the `Nuke` workflow with that cloud selected, with `dry_run` first to see what would go. Locally, with credentials in the shell:

```sh
git clone https://github.com/entigolabs/entigo-infralib-test ../entigo-infralib-test
../entigo-infralib-test/scripts/nuke.sh aws --dry-run
../entigo-infralib-test/scripts/nuke.sh oracle --prefix biz --dry-run
AZURE_TENANT_ID=... AZURE_SUBSCRIPTION_ID=... AZURE_CLIENT_ID=... AZURE_CLIENT_SECRET=... ../entigo-infralib-test/scripts/nuke.sh azure --dry-run
```

A resource that must survive the nuke goes into the filters of its cloud's file. The nuke keeps no module, state bucket or cluster on purpose.
