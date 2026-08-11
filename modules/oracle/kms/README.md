## Opinionated module for OCI Vault key creation ##

Creates one vault and the same three keys the `aws/kms` and `google/kms` modules create -
`data`, `config` and `telemetry` - so that a deployment describes its encryption the same
way on every cloud. Other modules take the OCIDs from this module's outputs.

Unlike AWS there is no key policy attached to a key: OCI authorises key use through IAM
policies written against the compartment, so grants live with the consumer, not here.

### The CA key ###

A fourth key, `ca`, is created by default. It is the signing key for the certificate
authority in `modules/oracle/pca`, which issues the wildcard certificate every app behind
the native ingress controller is served from.

It lives here rather than in that module because this module owns the vault, and a second
vault would bring a 7-day deletion floor of its own. The IAM grant that lets an authority
*use* this key is the other way round - it belongs to `oracle/pca`, because the module that
creates the CA is the one that has to wait for the grant to propagate.

It is the only **HSM-protected** key in the module, and that is not a choice: OCI
Certificates refuses software-protected keys for a certificate authority ("Certificates
doesn't support the use of software-protected keys" - Oracle's own wording), and the
Console key picker lists HSM asymmetric keys only. HSM key versions are billed per version,
software ones are free, so this is the only line in the module that costs money.

Note what it is *not*: the vault is a `DEFAULT` vault, using OCI's shared HSM partitions.
A **Virtual Private Vault** gives dedicated partitions and is billed per hour whether or
not it holds a key - `vault_type` can select it, but nothing here needs it.

Set `create_ca_key = false` in a deployment that issues no certificates from an OCI CA.

### Deletion is always scheduled ###

Vaults and keys cannot be deleted immediately - OCI schedules deletion 7 to 30 days out and
the object keeps its name until then. Every name this module creates therefore carries a
random suffix, so that tearing a deployment down and rebuilding it the same day does not
collide with its own pending-deletion leftovers.

### Example code ###

```
    modules:
      - name: kms
        source: oracle/kms
        inputs:
          compartment_id: '{{ .agent.accountId }}'
```
