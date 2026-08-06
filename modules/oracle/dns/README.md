## Oracle DNS zone and wildcard certificate ##

Creates the OCI DNS zone for `<subdomain_name>.<parent_domain>` and, like
`modules/aws/route53` does with ACM, a wildcard certificate covering it. The parent zone
lives elsewhere (typically Route53) and the NS delegation is added by hand from the
`name_servers` output.

### The certificate is issued by an internal CA ###

`*.<domain>` plus the apex, issued by a root CA this module creates in OCI Certificates
and signs with an HSM key from `modules/oracle/kms` (`ca_key_id`, wired automatically).

A wildcard rather than per-app certificates because
`modules/k8s/oci-native-ingress-controller` terminates TLS on an OCI load balancer
listener, and a listener holds exactly one key pair - there is no SNI, so every app on the
cluster is served from this one certificate, referenced by OCID.

An internal CA rather than an imported public certificate because OCI has no equivalent of
ACM's free publicly trusted issuance, and an imported certificate cannot be renewed without
a human: `oci_certificates_management_certificate` rejects IMPORTED material that is unknown
at plan time (see the comment on `certificate_ocid` in `variables.tf`), so terraform can
never own one. A CA-issued certificate is renewed by OCI itself through `certificate_rules`,
which is what makes this the same arrangement `modules/aws/route53` has with ACM and
`modules/google/dns` has with Certificate Manager.

**These certificates are trusted only by clients that have imported the CA.** Browsers, CI
jobs, `curl` and the argocd CLI all need it. Terraform cannot output the PEM - it comes
from the Certificates *data plane* API and the OCI provider only wraps
`certificates_management` - so fetch it with the CLI:

```shell
oci certificates certificate-authority-bundle get \
  --certificate-authority-id "$(terraform output -raw certificate_authority_id)" \
  --query 'data."certificate-pem"' --raw-output > ca.pem
```

To use a publicly trusted certificate instead, issue and import it outside terraform and
pass its OCID as `certificate_ocid`; the CA and wildcard are then not created and the OCID
is passed straight through.

### Renewal has a known gap ###

OCI renews the certificate on `certificate_renewal_interval`, producing a new *version* of
the same OCID. An OCI load balancer listener has been observed to keep serving the previous
version until the ingress controller re-pushes it, so until that is confirmed to behave
differently for CA-issued certificates, follow a renewal with:

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
      - name: dns
        source: oracle/dns
        inputs:
          compartment_id: '{{ .agent.accountId }}'
          parent_domain: "example.com"
```
