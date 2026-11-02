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
unexpected state. This module creates that policy, then **waits 300 seconds**
before creating the CA, because IAM is eventually consistent and a CA that starts too early
fails permanently instead of retrying. 60s was tried and was not enough.

That wait is why this module is worth applying once and leaving alone.

### Subordinate CAs ###

Setting `issuer_certificate_authority_id` makes the CA subordinate rather than self-signed,
which is how a deployment chains to a root held centrally: the root lives in another
compartment or tenancy, this module creates an intermediate in the deployment's own
compartment, and clients that already trust the root need no new import.

**Adopting a subordinate is tested.** The biz and pri tests adopt `biz-net-pca-ca` and
`pri-net-pca-ca` with `create_ca = false` and `ca_name`. Both are subordinates of a root in the
same compartment, `dns` issues its certificates from them, and clients that import only the
root trust those certificates.

**Creating one works, but not with the defaults.** OCI rejects a subordinate when:

* its validity ends after its issuer's - `ca_validity_years = 10` does, under any 10-year root
  older than a few days: "The ending date of the validity period … cannot exceed the ending
  date of the validity period … of the CA.";
* its validity is longer than its issuer's `subordinate_ca_max_validity` - `P1095D` for a root
  this module creates: "The validity period 1460 days … exceeds the maximum validity period
  allowed 1095 days …";
* its own `subordinate_ca_max_validity` is longer than its validity - the default `P1095D` is,
  on a 2-year subordinate: "The maximum validity duration 1095 days … exceeds the certificate
  authority validity period 730 days …".

`ca_validity_years = 2` with `subordinate_ca_max_validity = "P365D"` passes all three. A
subordinate may share its issuer's `ca_key_id`.

**An issuer in another compartment of the same tenancy needs no grant on its side.** ei-agent's
DevOps build pipelines, which may manage resources only in the deployment's own compartment,
created a subordinate of a root in another compartment, and its chain verifies against that
root. OCI checks only the subordinate's own compartment: `CreateCertificateAuthority` needs
`manage certificate-authorities` and `use certificate-authority-delegates` there, and nothing on
the issuer ([Details for Certificates](https://docs.oracle.com/en-us/iaas/Content/Identity/Reference/certificatespolicyreference.htm)).
So a central root does not limit which compartments chain to it: anything that may create a CA
in any compartment of the tenancy can get one issued by it. An issuer in another tenancy is
untested.

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
**two passes a day apart**.

`name_salt` appends a random suffix to the CA name so a rebuild cannot collide with its own
leftovers. It is **off by default** - a deployment built once deserves a readable name - and
worth turning on for an environment that is torn down repeatedly. The suffix lives in
terraform state, so a nuke that takes the state bucket with it produces a fresh one on the
next run, with nothing to bump by hand.

### Example code ###

```
    modules:
      - name: kms
        source: oracle/kms
      - name: pca
        source: oracle/pca
        inputs:
          organization: "Entigo AS"
          country: "EE"
      - name: dns
        source: oracle/dns
        inputs:
          domains: |
            {
              "public" = {
                domain_name = "dev.example.com"
              }
            }
```

`ca_key_id` is wired from `kms` by `agent_input.yaml` and `certificate_authority_id` into
`dns` the same way, so neither appears in the config. To use a CA held elsewhere, leave this
module out and set `certificate_authority_id` on `dns` explicitly.
