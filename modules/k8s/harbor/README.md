## Opinionated helm package for harbor ##

Installs Harbor and, with `global.createProxyRepos`, a job that sets it up as a
pull-through cache for Docker Hub, ghcr.io, gcr.io and registry.k8s.io.

### Admin password and secret key ###

The chart defaults (`Harbor12345` and `not-a-secure-key`) are never used.
With `global.generateAdminSecret: true`, the default, External Secrets generates
both once into the Secret `harbor-admin` (`templates/adminSecret.yaml`):

| Key | Used for |
|---|---|
| `HARBOR_ADMIN_PASSWORD` | the `admin` user, 32 characters, also read by the proxy job |
| `secretKey` | encryption of stored credentials, 16 characters |

This needs the `external-secrets` module in the same platform. The values are
generated with `refreshPolicy: CreatedOnce` and are never rotated: Harbor reads
the admin password only when it creates its database, and changing `secretKey`
makes stored credentials unreadable.

Read the password with:

```
kubectl -n <module name> get secret harbor-admin -o jsonpath='{.data.HARBOR_ADMIN_PASSWORD}' | base64 -d
```

### Upgrading an existing install ###

An install from before this change still has the admin password it was created
with (`Harbor12345` unless changed) and `secretKey: not-a-secure-key`. A newly
generated Secret would not match either, so the proxy job fails to log in and
stored credentials can no longer be decrypted. Before upgrading:

1. Change the admin password in the Harbor UI, or with
   `PUT /api/v2.0/users/1/password`.
2. Create the Secret yourself with the current values:

   ```
   kubectl -n <module name> create secret generic harbor-admin \
     --from-literal=HARBOR_ADMIN_PASSWORD='<new admin password>' \
     --from-literal=secretKey='not-a-secure-key'
   ```

3. Turn generation off so the Secret is only read:

   ```yaml
   inputs:
     global:
       generateAdminSecret: false
   ```

To use a Secret with another name, set `global.generateAdminSecret: false` and
point `harbor.existingSecretAdminPassword`, `harbor.existingSecretAdminPasswordKey`
and `harbor.existingSecretSecretKey` at it. The secret key must be stored under
the key `secretKey`.

### Example code ###

```
    modules:
      - name: harbor
        source: harbor
```
