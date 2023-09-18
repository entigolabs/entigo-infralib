## Helm charts that we use ##

These modules can be used in the [entigo-infralib-agent](https://github.com/entigolabs/entigo-infralib-agent) steps of "__type: argocd-apps__". They will be launched using ArgoCD, but also "aws/helm-git" TF module could be used to install these without ArgoCD.

## Example code ##
```
steps:
  - name: applications
    type: argocd-apps
    workspace: test
    version: stable
    vpc_prefix: network-vpc
    argocd_prefix: infrastructure-argocd
    eks_prefix: infrastructure-eks
    modules:
      - name: hello-world
        source: hello-world
        version: stable

```
