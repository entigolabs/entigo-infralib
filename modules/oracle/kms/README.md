## Opinionated module for OCI Vault key creation ##

Creates one vault and the same three keys the `aws/kms` and `google/kms` modules create -
`data`, `config` and `telemetry` - so that a deployment describes its encryption the same
way on every cloud. Other modules take the OCIDs from this module's outputs.

Unlike AWS there is no key policy attached to a key: OCI authorises key use through IAM
policies written against the compartment, so a key carries no document of its own.

### Who uses which key ###

Same split as `aws/kms` and `google/kms`, so a deployment's encryption reads the same everywhere:

| key | consumer | wired by |
|---|---|---|
| `telemetry` | loki's chunk bucket | `modules/k8s/loki` — automatic |
| `data` | worker node boot volumes | `modules/oracle/oke`'s `node_kms_key_id` — automatic |
| `config` | OKE etcd, and so every Kubernetes Secret | `oke`'s `etcd_kms_key_id` — **opt-in** |
| `ca` | the certificate authority | `modules/oracle/pca` — automatic |

`etcd_kms_key_id` is the one you have to ask for. OCI cannot re-key an existing cluster, so it
is creation-time only: setting it on a live cluster makes terraform plan a **replacement**.
Since the agent applies plans unattended, wiring it automatically would turn "add a kms module"
into "rebuild the cluster", so it is left to a deployment to set explicitly on a cluster that
does not exist yet.

Block volumes behind PersistentVolumeClaims are **not** covered yet - OKE's default `oci-bv`
StorageClass names no key, so monitoring PVCs still use Oracle-managed encryption. That needs a
StorageClass carrying `kmsKeyId`, which no module creates today.

### The service grants ###

A service encrypting on your behalf does so as *itself*, and is refused unless a policy allows
it - the same failure shape `modules/oracle/pca` hits with its CA, and it does not look like a
permission problem: OKE simply refuses to create a cluster whose etcd key it cannot read, and a
bucket or volume naming an unusable key is rejected at create time.

So this module grants `use keys in compartment` to `oke` and `blockstorage`, plus
`objectstorage-<region>` - Object Storage's principal is the one that carries a region, which is
why the module takes `region` (wired from `{{ .agent.region }}`). The three key outputs then
depend on a propagation wait, so every consumer inherits it and nothing encrypts with a key
before the grant lands.

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
