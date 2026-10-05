## Customer managed keys ##

Creates an RBAC Key Vault (soft delete + purge protection, required by disk encryption sets and storage CMK) with
`data`, `config` and `telemetry` RSA keys, like aws/kms and google/kms: `data` for disks and other data (azure/aks
encrypts node and PVC disks with it), `telemetry` for monitoring data (loki, mimir), `config` for customer app secrets.

`*_key_id` outputs are versionless and follow key rotation, like the aws alias ARNs and google key ids (Key Vault has
no aliases, the key name is the stable reference). `*_key_version_id` pins the current version, `*_key_resource_id`
is the scope for role assignments on the key.

Azure has no service wide key access like aws key policies or google service agents: every consumer uses its own
managed identity, so consumer modules assign it a role on the key (Key Vault Crypto Service Encryption User) and wait
for it to propagate. `*_key_users` grants key roles to additional principals (Crypto User on the config key, Crypto
Service Encryption User on the others). The keys only allow encrypt, decrypt, wrap and unwrap.

Keys are software protected by default (like google/kms and oracle/kms). The vault is Premium, which costs the same as
Standard for software keys, so HSM protected keys only need `key_type: "RSA-HSM"`. Set it at install time: changing
`key_type` replaces the keys.

The vault name is `<prefix, max 15>-<agent uniqueSuffix>`, the same on every run of an environment. With purge
protection a deleted vault and its keys stay recoverable and the name reserved for `soft_delete_retention_days`
(default 30, like the aws/kms deletion window). Re-creating the environment within that time recovers the deleted
vault with its keys (azurerm default), and the keys then fail with "already exists": delete the recovered keys
(`az keyvault key delete --vault-name <vault> -n <key>`) and run again, they are recovered as soft-deleted keys. The vault is RBAC mode: Owner and
Contributor don't give access to its keys, admins need a Key Vault data role. The identity that installs the module
(normally the agent job identity) gets Key Vault Crypto Officer and keeps it: a later run by someone else doesn't
replace it, so a local run can't remove the agent's key access. Everyone else who runs the module (people, CI) needs
the role too, because the plan reads the keys: add them to `admin_object_ids`. Don't list the installing identity there
(the plan rejects it). If the installing identity is deleted and recreated, it gets a new object
id and the plan fails with 403 on the keys; an Owner restores the role:
```
az role assignment create --assignee <object id> --role "Key Vault Crypto Officer" --scope <kms key_vault_id>
```
Applying the module needs `Microsoft.Authorization/roleAssignments/write`
on the resource group (Owner, User Access Administrator or Role Based Access Control Administrator; the agent job
identity is Owner), Contributor alone fails with an authorization error.

`delete_lock_enabled` (default on) adds a CanNotDelete lock on the vault, creating it needs Owner or User Access
Administrator. The lock stops deleting the vault or its resource group in the portal, CLI or another tool, but not a
Terraform destroy or replacement of the vault: Terraform deletes the lock first (unlike google/kms `prevent_destroy`).
Deleted vaults and keys can still be recovered within the soft delete retention thanks to purge protection.

### Example code ###
```
    modules:
      - name: kms
        source: azure/kms
        inputs:
          key_rotation_period: "P90D"
```
