## Bootstrap argocd with terraform ##

This is a specialized version of the __helm-git__ module. 

If we do not have a Helm chart registry then we can't just install our k8s modules with only [helm_release](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release) terraform resource. 
This module will clone the desired helm chart from git and install it using __helm_release__.

__repository__ the repository to use for cloning (https://github.com/entigolabs/entigo-infralib-release)

__branch__ the branch to use (main for latest version in our repo).

__name__ how to name the helm release.

__namespace__ what namespace to use.

__create_namespace__ defaults to true, if set to false the namespace has to exist or no namespaced resources are created.

__argocd_apps_name__ the prefix + application step name to bootstrap the app-of-apps from, defaults to "ep-applications".

__install_crd__ Defaults to true, set to false to skip Argocd CRD intallation.

__ingress_group_name__ Defaults to "internal", set to "external" if you want Argocd to be on public internet.

__ingress_scheme__ Defaults to "internal", set to "internet-facing" if you want Argocd to be on public internet.

### SSM parameters ###
"/entigo-infralib/${local.hname}/repo_url"


### Example code ###

```
    modules:
      - name: argocd
        source: aws/argocd
        inputs:
          depends_on: |
            [module.eks]
          hostname: "argocd.ep-infrastructure-dns-test.infralib.entigo.io"
          ingress_group_name: "external"
          ingress_scheme: "internet-facing"
          branch: "main"
          namespace: "argocd"
          name: "argocd"
          argocd_apps_name: "ep-applications"

```
### Limitations ###
If you want multiple ArgoCD instances then they have to be under different workspaces. It uses by default the kubernetes from the same workspace.
