# Extra CA certificates

Every `*.crt` here is handed to the agent of every environment
(`agents/<env>/ca-certificates/`, installed into the trust store of each step)
and mounted into the route probes of the tests. A certificate one environment
alone needs goes to `environments/<env>/ca-certificates/` instead.

- `letsencrypt-stg-root-x1.crt`, `letsencrypt-stg-root-x2.crt`: the Let's
  Encrypt staging roots, from https://letsencrypt.org/docs/staging-environment/.
  The Azure environments issue their gateway certificates from the staging
  server (`modules/k8s/cert-manager/test/azure_*.yaml`), so the agent's `argocd`
  CLI would not trust ArgoCD there otherwise.
- `infralib-root-ca.crt`: the root of the OCI certificate authorities the
  Oracle environments adopt through `modules/oracle/pca`; the dns module
  issues the gateway certificates from them. Exported with
  `oci certificates certificate-authority-bundle get`, see
  `modules/oracle/pca`'s `certificate_authority_bundle_command` output.
