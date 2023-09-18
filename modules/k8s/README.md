## Helm charts that we use ##

These modules can be used in the entigo-agent steps of "__type: argocd-apps__". They will be launched using argocd by default but also "aws/helm-git" module could be used to invoke them without ArgoCD.

## Example code ##
```
steps:
  - name: network
    type: terraform
    workspace: test
    approve: minor
    modules:
      - name: hello
        source: aws/hello-world
        version: stable

```
