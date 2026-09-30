## Customer managed keys ##

Creates an RBAC Key Vault (soft delete + purge protection, required by disk encryption sets and storage CMK) with
`data`, `config` and `telemetry` RSA keys, like aws/kms and google/kms. Consumers use the `*_key_versionless_id`
outputs (follow key rotation) or the versioned `*_key_id`. azure/aks encrypts node and PVC disks with the data key.

Azure has no service wide key access like aws key policies or google service agents: every consumer uses its own
managed identity, so consumer modules assign it a role on the key (Key Vault Crypto Service Encryption User) and wait
for it to propagate. `*_key_users` grants key roles to additional principals. The keys only allow encrypt, decrypt,
wrap and unwrap.

Keys are software protected by default (like google/kms and oracle/kms). The vault is Premium, which costs the same as
Standard for software keys, so HSM protected keys only need `key_type: "RSA-HSM"`. Set it at install time: changing
`key_type` replaces the keys.

With purge protection a deleted vault and its keys stay recoverable and the name reserved for
`soft_delete_retention_days` (default 30, like the aws/kms deletion window). The vault is RBAC mode: Owner and
Contributor don't give access to its keys, admins need a Key Vault data role (`admin_object_ids`, the caller is always
added). `delete_lock_enabled` (default on) adds a CanNotDelete lock on the vault, creating it needs Owner or User Access
Administrator.

### Example code ###
```
    modules:
      - name: kms
        source: azure/kms
        inputs:
          key_rotation_period: "P90D"
```
