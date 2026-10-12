## Deprecated

This module is no longer actively developed. Use the `k8s/external-secrets`
module instead, which does the same job with an `ECRAuthorizationToken`
generator and needs no extra workload in the cluster.

`argocd-ecr-updater` runs a Deployment that watches the ArgoCD repository
Secrets and rewrites their password with a fresh ECR token.
`k8s/external-secrets` produces the same credentials declaratively: an
`ECRAuthorizationToken` generator per ECR account and region, and an
`ExternalSecret` that renders it into an ArgoCD credentials Secret and refreshes
it every 30 minutes. It also creates the IAM role behind it, with the same ECR
permissions this module's role had.

## Migration from k8s/argocd-ecr-updater to k8s/external-secrets

1. Make sure `external-secrets` is in the same `argocd-apps` step. On AWS it
   already sets the ECR credentials up by default — the agent fills
   `global.aws.createECRNamespace` with the ArgoCD module's namespace and seeds
   `global.aws.createECRAccounts` with the cluster's own account and region.

   Only if you pull from an ECR registry in another account or region, list
   every one of them explicitly, because setting the input replaces the
   default rather than adding to it:

   ```yaml
      - name: apps
        type: argocd-apps
        modules:
            ...
            - name: external-secrets
              source: external-secrets
              inputs:
                global:
                  aws:
                    createECRAccounts:
                      - accountNumber: "111111111111"
                        accountRegion: "eu-north-1"
                      - accountNumber: "222222222222"
                        accountRegion: "us-east-1"
            ...
   ```

2. Run the infralib agent. Verify that the pipeline is successful.

3. Verify the credentials exist. For every account and region there should be a
   `repo-<accountNumber>-<accountRegion>` Secret in the ArgoCD namespace,
   labelled `argocd.argoproj.io/secret-type: repo-creds`:

   ```bash
   kubectl -n <argocd namespace> get secret -l argocd.argoproj.io/secret-type=repo-creds
   ```

   The ArgoCD UI lists them under Settings -> Repository certificates and known
   hosts -> Credential templates. Confirm an Application that pulls a Helm
   chart from ECR still syncs.

4. Remove argocd-ecr-updater from the agent configuration `argocd-apps` step.

   Before:

   ```yaml
      - name: apps
        type: argocd-apps
        modules:
            ...
            - name: argocd-ecr-updater
              source: argocd-ecr-updater
            ...
   ```

   After: the entry is gone, nothing replaces it.

5. Run the infralib agent again. Verify that the pipeline is successful.

6. Delete the argocd-ecr-updater app in the ArgoCD UI.

7. Delete the argocd-ecr-updater namespace with
   `kubectl delete namespace argocd-ecr-updater`.

8. Optional. This module patched ArgoCD `repository` Secrets, so every ECR
   repository needed its own entry. The replacement creates `repo-creds`
   instead, which apply to every repository URL under the registry host, so any
   per-repository Secret that existed only to carry an ECR password can be
   deleted.
