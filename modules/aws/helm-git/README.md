## Install one of our k8s modules using terraform ##
If we do not have a Helm chart registry then we can't just install our k8s modules with only [helm_release](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release) terraform resource. 
This module will clone the desired helm chart from git and install it using __helm_release__.

__repository__ the repository to use for cloning (https://github.com/entigolabs/entigo-infralib-release)

__branch__ the branch to use (main for latest version in our repo).

__path__ the git path to the helm chart (In our repo it starts with "modules/k8s/...")

__values__ the helm values to pass to the helm package (multi line string with yaml inside).

__name__ how to name the helm release.

__namespace__ what namespace to use.

__create_namespace__ defaults to true, if set to false the namespace has to exist or no namespaced resources are created.




### Example code ###

```
    modules:
      - name: argocd
        source: aws/helm-git
        inputs:
          depends_on: |
            [module.eks]
          branch: "main"
          path: "modules/k8s/argocd"
          name: "argocd"
          values: |
            argocd:
              crds:
                install: false

```
