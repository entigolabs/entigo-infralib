## Oppinionated module for eks creation ##


Oppinionated version of this https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest

__eks_oidc_provider__ OIDC of EKS
__eks_oidc_provider_arn__ OIDC arn of EKS
__region__ - region where EKS is installed
__account__ - account number where EKS is installed

### Example code ###

```
    modules:
      - name: crossplane
        source: aws/crossplane
        inputs:
          eks_prefix: ep-infrastructure-eks

```

