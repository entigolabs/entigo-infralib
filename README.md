# entigo-infralib


Usually we release once per day. During the evening the entigo-infralib AWS account is nuked(Nuke action). In the morning latest release is installed and tests executed(Stable action) and after that it is upgraded to "main" branch and tests are executed. If the tests are passed and main is not the same state as last release, then a new release is created.

Once the release is created then we make it public in [entigo-infralib-release](https://github.com/entigolabs/entigo-infralib-release) repository. The released modules and profiles can then be used by [entigo-infralib-agent](https://github.com/entigolabs/entigo-infralib-agent) or called directly from terraform code or argocd applications.

To create a new release the "Release" action should be run. It will create a release if main branch differs from last release.


## Folders ##

__modules__ contains opinnionated terraform modules or kubernetes helm charts that we repeatedly use in our projects.

__images__ contains the runtime images for running infrastructure as code.

__profiles__ contains base profiles used by entigo-infralib-agent that we use on many clients. It is a way of combining different module with common inputs.

__providers__ contians provider configurations for terraform modules.


## Example code ##
Terraform code example:
```
module "vpc" {
  source                 = "git::https://github.com/entigolabs/entigo-infralib-release.git//modules/aws/vpc?ref=v0.6.3"
  prefix                 = "ep-network-vpc"
  one_nat_gateway_per_az = true
}

```
K8S example argocd application:
```
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: istio-base
spec:
  destination:
    server: https://kubernetes.default.svc
    namespace: istio-base
  project: default
  sources:
    - repoURL: 'ssh://your/repo'
      targetRevision: 'main'
      ref: codeRepo
    - repoURL: "https://github.com/entigolabs/entigo-infralib-release.git"
      targetRevision: 'v0.6.3'
      path: "modules/k8s/istio-base"
      helm:
        ignoreMissingValueFiles: true
        valueFiles:
          - 'values.yaml'
          - '$codeRepo/ep-applications/test/istio-base/values.yaml'
  syncPolicy:
    automated:
      selfHeal: true
    syncOptions:
      - CreateNamespace=true

```
