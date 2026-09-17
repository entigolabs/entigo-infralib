# crossplane-oracle

Installs the Oracle-official Crossplane OCI provider
([oracle/crossplane-provider-oci](https://github.com/oracle/crossplane-provider-oci),
packages on ghcr.io) with OKE Workload Identity authentication. Mirrors crossplane-aws /
crossplane-google; depends on crossplane-core (Crossplane v2 - the provider's v1.x line
requires it).

Per-service providers are selected via `global.requiredProviders` /
`global.extraProviders` (`identity` and `objectstorage` by default). Consuming modules
create their own OCI IAM policies with `identity.oci.upbound.io/v1alpha1 Policy` CRs in
`templates/oracle/` - see external-dns and cluster-autoscaler for the pattern.

Auth notes:

- OKE Workload Identity, so the provider acts as its own service account rather than as
  every instance in the compartment. Requires an ENHANCED cluster, which modules/oracle/oke
  creates. The service account name is pinned in deploymentRuntimeConfig.yaml rather than
  generated per provider revision, which is what lets terraform name it in the bootstrap
  policy before the cluster has any workloads.
- The provider needs a Terraform-created bootstrap grant to manage the per-app policies:
  `manage policies in compartment` on the shared dynamic group - created by
  modules/oracle/oke. Statements scoped `in tenancy` cannot live in compartment-attached
  policies; a module needing them would also need the bootstrap grant widened to tenancy.
- The credentials secret carries only tenancy OCID + region (no secret material).
- Provider issue #123 (sensitive fields insufficiently protected) - review before
  storing production credentials through this provider.
