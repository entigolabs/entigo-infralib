## Opinionated module for an OCI private certificate authority ##

Creates one certificate authority in OCI Certificates and the IAM grant it needs to work.
Nothing else - the certificates it signs belong to the module that owns the domain, which is
`modules/oracle/dns`. It takes `certificate_authority_id` from here the way
`modules/aws-v2/route53` takes `certificate_authority_arn` from an AWS Private CA.

The split exists so that the CA can outlive, and sit apart from, the deployments that use it.
A CA is a long-lived object with a 7-day deletion floor whose whole value is that clients have
imported it; a DNS zone is rebuilt whenever a deployment is. Keeping them in one module ties
the two lifecycles together and offers no way to point several deployments at one authority.

### Why a private CA at all ###

OCI has no equivalent of ACM's or Certificate Manager's free publicly trusted issuance. The
`config_type` enum in the provider is the whole story: `IMPORTED`, `ISSUED_BY_INTERNAL_CA`,
`MANAGED_EXTERNALLY_ISSUED_BY_INTERNAL_CA`. Every path is either "you brought the PEM" or "our
internal CA signed it", and there is no DNS-validation subsystem anywhere in the API.

An imported certificate is not a workaround, because terraform cannot own one:
`oci_certificates_management_certificate` validates IMPORTED material in `CustomizeDiff` and
rejects anything unknown at plan time, so a certificate produced in the same apply always
fails. Renewal hits the same check. A CA-issued certificate is renewed by OCI itself, which is
what makes this the same arrangement AWS has with ACM and GCP with Certificate Manager.

**Certificates from this CA are trusted only by clients that have imported it** - browsers, CI
jobs, `curl`, the argocd CLI. Terraform cannot output the PEM: it comes from the Certificates
*data plane* API and the provider only wraps `certificates_management`. The
`certificate_authority_bundle_command` output is the CLI invocation that exports it.

### The signing key ###

`ca_key_id` must be an **HSM-protected** asymmetric key, wired from `modules/oracle/kms`
automatically. This is not a preference: OCI Certificates rejects software-protected keys
outright ("Certificates doesn't support the use of software-protected keys" - Oracle's wording)
and the Console key picker lists HSM keys only. HSM key versions are billed per version, so
that key is the one line of a deployment's encryption setup that costs money.

### The grant, and the wait ###

A CA reaches for its signing key *as itself*, and is refused unless a policy allows it. The
symptom is bad: the CA is created, goes to FAILED seconds later, and the only explanation is in
the Console - `lifecycle-details` is empty over the API and terraform reports nothing but an
unexpected state. This module creates the dynamic group and policy, then **waits 300 seconds**
before creating the CA, because IAM is eventually consistent and a CA that starts too early
fails permanently instead of retrying. 60s was tried and was not enough.

That wait is why this module is worth applying once and leaving alone.

### Subordinate CAs ###

Setting `issuer_certificate_authority_id` makes the CA subordinate rather than self-signed,
which is how a deployment chains to a root held centrally: the root lives in another
compartment or tenancy, this module creates an intermediate in the deployment's own
compartment, and clients that already trust the root need no new import.

**Untested.** Every CA this repo has created has been a self-signed root, and the issuing CA
must separately grant this compartment the use of it - not something the subordinate's side can
arrange. Treat it as a starting point, not a supported path.

### Deliberately not here ###

**Revocation.** `certificate_revocation_list_details` is not wired. A CRL needs an Object
Storage bucket to publish to, and nothing in infralib consumes one - the certificates this CA
issues are wildcards on a load balancer listener, replaced by renewal rather than revoked. Add
it when a client requires it, not before.

**Tags.** OCI has no provider-level `default_tags`, and tagging is applied through a Tag Default
Rule on the compartment rather than per resource, so no module here sets them.

### Deletion is always scheduled ###

A CA cannot be deleted immediately - OCI schedules deletion 7 days out at the earliest and the
name is held for the whole period. Worse, a CA cannot be scheduled at all while any certificate
it issued still exists, and a certificate's own floor is 24 hours, so a full teardown takes
**two passes a day apart**. Every name here therefore carries a random suffix; see `name_salt`
for the failure this rescues.

### Example code ###

```
    modules:
      - name: kms
        source: oracle/kms
      - name: pca
        source: oracle/pca
        inputs:
          compartment_id: '{{ .agent.accountId }}'
          organization: "Entigo AS"
          country: "EE"
      - name: dns
        source: oracle/dns
        inputs:
          compartment_id: '{{ .agent.accountId }}'
          parent_domain: "example.entigo.dev"
```

`ca_key_id` is wired from `kms` by `agent_input.yaml` and `certificate_authority_id` into
`dns` the same way, so neither appears in the config. To use a CA held elsewhere, leave this
module out and set `certificate_authority_id` on `dns` explicitly.
