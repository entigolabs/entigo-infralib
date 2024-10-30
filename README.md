# entigo-infralib


Usually we release once per day. During the evening the entigo-infralib AWS and Google Cloud accounts are nuked(Nuke action). In the morning latest release is installed and tests executed(Stable action) and after that it is upgraded to "main" branch and tests are executed. If the tests are passed and main is not the same state as last release, then a new release is created.

Once the release is created then we make it public in [entigo-infralib-release](https://github.com/entigolabs/entigo-infralib-release) repository. The released modules can then be used by [entigo-infralib-agent](https://github.com/entigolabs/entigo-infralib-agent) or called directly from terraform code or ArgoCD applications.

To create a new release the "Release" action should be run. It will create a release if main branch differs from last release.


## Folders ##

__modules__ contains opinnionated Terraform modules or Kubernetes Helm charts that we repeatedly use in our projects.

__images__ contains the runtime images for running infrastructure as code.

__providers__ contians provider configurations for Terraform modules.


## Example code ##
Terraform code example:
```
module "main" {
  source                 = "git::https://github.com/entigolabs/entigo-infralib-release.git//modules/aws/vpc?ref=v1.0.14"
  prefix                 = "dev-net-main"
  elasticache_subnets    = []
  intra_subnets          = []
  one_nat_gateway_per_az = false
  vpc_cidr               = "10.112.0.0/16"
}

```
ArgoCD Application example:
```
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: 'external-dns-dev'
spec:
  destination:
    server: https://kubernetes.default.svc
    namespace: 'external-dns-dev'
  project: default
  sources:
    - repoURL: 'https://github.com/entigolabs/entigo-infralib-release.git'
      targetRevision: 'v1.0.14'
      path: "modules/k8s/external-dns"
      helm:
        ignoreMissingValueFiles: true
        valueFiles:
          - 'values.yaml'
          - 'values-aws.yaml'
        values: |
          global:
              aws:
                  account: "XXXX"
                  clusterOIDC: oidc.eks.eu-north-1.amazonaws.com/id/XXXX
          
  syncPolicy:
    syncOptions:
      - CreateNamespace=true
      - RespectIgnoreDifferences=true

```
