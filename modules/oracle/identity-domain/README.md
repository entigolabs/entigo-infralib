## Opinionated module for an OCI compartment-scoped identity domain ##

Creates one identity domain in a compartment. Unlike classic/default-domain users which are
tenancy-scoped, domain-hosted users and groups live in the domain's home compartment, enabling
IAM grants narrowed from `manage users in tenancy` to `manage users in compartment id X` -
allowing compartment-scoped automation without tenancy-admin grants.

### Use case ###

Loki uses this for its Object Storage access. The S3-compatibility endpoint authenticates
solely with a Customer Secret Key, which belongs to an IAM user - requiring Crossplane to
create that user. Without an identity domain, that user would have to be tenancy-scoped,
forcing the grant to be tenancy-scoped too, and thus requiring tenancy-admin approval.

With this domain, the user is domain-scoped (compartment-scoped), the grant can be narrowed
to the compartment, and compartment-level automation can proceed without tenancy intervention.

### The domain ###

This module creates only the domain itself with `license_type = "FREE"`. The domain has no human
administrator and no hosted authentication - it exists purely as a container for machine-to-machine
users that Crossplane creates on demand.

### Name collisions ###

An identity domain cannot be deleted immediately - OCI schedules deletion for a retention period.
Rebuilds that collide with pending-deletion leftovers fail. `name_salt` appends a random suffix
to prevent this, and is **off by default** - a deployment built once deserves a readable name,
and worth turning on for environments that are torn down repeatedly. The suffix lives in
terraform state, so a nuke that takes the state bucket with it produces a fresh one on the
next run.

### Example code ###

```
    modules:
      - name: identity-domain
        source: oracle/identity-domain
        inputs:
          compartment_id: '{{ .agent.accountId }}'
```

`idcs_endpoint` and `display_name` are wired from this module to k8s/loki by `agent_input_oracle.yaml`.
