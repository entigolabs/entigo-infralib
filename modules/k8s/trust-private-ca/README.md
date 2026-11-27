# trust-private-ca

Lets the pods of the listed namespaces trust private CA certificates, for example the CA that signs the
gateway certificates of a platform without public certificates.

The certificates are kept in a ConfigMap in the module's namespace. Two Kyverno policies copy it into every
listed namespace and mount it into every pod created there, as `/etc/ssl/certs/<module name>.crt`. `mountPath`
sets another directory. Needs the `kyverno` module.

```yaml
    modules:
      - name: trust-private-ca
        source: trust-private-ca
        inputs:
          certificates:
            company-root: |
              -----BEGIN CERTIFICATE-----
              ...
              -----END CERTIFICATE-----
          namespaces:
            - argocd
            - grafana
```

- A pod gets the certificates when it is created: restart the pods that were already running.
- Programs that read the system directory `/etc/ssl/certs` trust them (Go, curl and OpenSSL on most images). A
  program with a certificate bundle of its own (Java, Python `certifi`, Node.js) needs its own setting.
- Without `certificates` or without `namespaces` the module installs nothing.
