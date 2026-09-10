## Oracle DNS zones and certificates ##

Creates OCI DNS zones and a wildcard certificate for each, from one `domains` map - the same
input shape as `modules/aws-v2/route53`, so a deployment reads the same on both clouds.
Certificates are issued by the authority in `modules/oracle/pca`, wired in automatically.

### The domains map ###

Each key names a domain. Everything but `domain_name` is optional:

| field | default | meaning |
|---|---|---|
| `domain_name` | - | the zone's name, e.g. `dev.example.com` |
| `create_zone` | `true` | false looks the zone up by name instead of creating it |
| `parent_zone_id` | `""` | OCID of the parent zone **when it is also in OCI DNS**; the NS delegation is then created here |
| `private` | `false` | resolvable only inside a VCN |
| `vcn_id` | module `vcn_id` | which VCN, for a private zone |
| `view_id` | looked up | overrides the view a private zone goes into |
| `create_certificate` | `true` | issue a wildcard for this domain |
| `certificate_authority_id` | module-level | issue from a different authority than the rest |
| `certificate_ocid` | `""` | use an existing certificate instead of issuing one |
| `default_public` | computed | this domain is what the `pub_*` outputs describe |
| `default_private` | computed | this domain is what the `int_*` outputs describe |

`default_public` and `default_private` are only worth setting when there is more than one
candidate: a single domain is both, and a single domain of a kind is the default for that kind.
More than one of either is an error, reported by a precondition on `pub_zone_id`.

**Deviation from `aws-v2/route53`:** each kind falls back to the other, so a deployment with
only public domains still has usable `int_*` outputs. That module requires exactly one default
of each and rejects the configuration otherwise. Every Oracle app reads
`.toutput.<dns>.int_domain`, so the fallback is what lets a single-public-zone deployment work.

Outputs are the `pub_*`/`int_*` scalars (`_zone_id`, `_domain`, `_cert_ocid`), the
`zone_ids` / `domain_names` / `certificate_ocids` / `nameservers` maps, and
`certificate_authority_id`.

### Private zones are views, not attachments ###

A Route53 private zone is attached to a VPC. An OCI private zone belongs to a DNS **view**, and
the view a VCN resolves against is the *default view of the resolver associated with that VCN* -
two hops from the VCN OCID a deployment actually has. This module makes that walk
(`oci_core_vcn_dns_resolver_association` → `oci_dns_resolver` → `default_view_id`), so a domain
only has to say `private = true` and inherit `vcn_id` from `modules/oracle/vpc`. Set `view_id`
to put a zone in a view of its own instead.

### The certificates are issued by a private CA ###

`*.<domain>` plus the apex, issued by the authority in `certificate_authority_id` - wired from
`modules/oracle/pca`, which owns the CA the way an AWS Private CA sits outside
`modules/aws-v2/route53` and is passed in as `certificate_authority_arn`. Unlike that module,
this one takes the authority at module level rather than per domain, because it is templated
from another module's output; a domain may still override it.

One wildcard per domain rather than per app because
`modules/k8s/oci-native-ingress-controller` terminates TLS on an OCI load balancer listener,
and a listener holds exactly one key pair - there is no SNI, so every app on a domain is served
from that domain's certificate, referenced by OCID.

A private CA rather than an imported public certificate because OCI has no equivalent of ACM's
free publicly trusted issuance, and an imported certificate cannot be renewed without a human:
`oci_certificates_management_certificate` rejects IMPORTED material that is unknown at plan time
(see `certificate_ocid` in `variables.tf`), so terraform can never own one. A CA-issued
certificate is renewed by OCI itself through `certificate_rules`, which is what makes this the
same arrangement `modules/aws-v2/route53` has with ACM and `modules/google/dns` has with
Certificate Manager.

**These certificates are trusted only by clients that have imported the CA.** Browsers, CI
jobs, `curl` and the argocd CLI all need it. Terraform cannot output the PEM - it comes from the
Certificates *data plane* API and the OCI provider only wraps `certificates_management` - so
`modules/oracle/pca` outputs the CLI command that exports it, as
`certificate_authority_bundle_command`.

To use a publicly trusted certificate instead, issue and import it outside terraform and pass
its OCID as that domain's `certificate_ocid`; nothing is issued for it and the OCID is passed
straight through to the outputs.

### There is no validation step ###

`aws-v2/route53` carries `create_validation`, DNS validation records, and a second public zone
whose only job is to answer an ACM challenge for a private domain. None of that exists here:
OCI Certificates has no validation subsystem, because a private CA signs what it is asked to.

### Renewal has a known gap ###

OCI renews each certificate on `certificate_renewal_interval`, producing a new *version* of the
same OCID. An OCI load balancer listener has been observed to keep serving the previous version
until the ingress controller re-pushes it, so until that is confirmed to behave differently for
CA-issued certificates, follow a renewal with:

```shell
kubectl rollout restart -n native-ingress-controller-system deploy/oci-native-ingress-controller
```

### Example code ###

```
    modules:
      - name: kms
        source: oracle/kms
        inputs:
          compartment_id: '{{ .agent.accountId }}'
      - name: pca
        source: oracle/pca
        inputs:
          compartment_id: '{{ .agent.accountId }}'
      - name: dns
        source: oracle/dns
        inputs:
          compartment_id: '{{ .agent.accountId }}'
          domains: |
            {
              "public" = {
                domain_name = "dev.example.com"
              },
              "private" = {
                domain_name = "int.example.com"
                private     = true
              }
            }
```

`certificate_authority_id` comes from `pca` and `vcn_id` from `vpc` through
`agent_input.yaml`, so neither appears in the config. To use a CA held centrally, leave `pca`
out and set `certificate_authority_id` explicitly.
