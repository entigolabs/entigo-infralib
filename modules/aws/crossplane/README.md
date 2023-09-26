## Oppinionated module for eks creation ##


Oppinionated version of this https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest

__eks_prefix__ look for OIDC based on this SSM parameter prefix ("/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/eks/.")


### Example code ###

```
    modules:
      - name: crossplane
        source: aws/crossplane
        inputs:
          eks_prefix: ep-infrastructure-eks

```

