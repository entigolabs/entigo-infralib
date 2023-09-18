## Oppinionated module for eks creation ##


Oppinionated version of this https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest

__vpc_prefix__ look for subnets based on this SSM parameter prefix ("/entigo-infralib/${var.vpc_prefix}-${terraform.workspace}/vpc/vpc_id")

__eks_cluster_version__ SHOULD NOT BE CHANGED, but can override the EKS version to use.

__iam_admin_role__ Defaults to "AWSReservedSSO_AdministratorAccess_.*" but could be something else depending on SSO setup.

__eks_cluster_public__ Defaults to false, if set true then the EKS cluster endpoint will be on the public internet

__eks_main_min_size__ Defaults to 2, minimum size of the main nodegroup. Set to 0 to disablet his nodegroup.

__eks_main_max_size__ Defaults to 3, maximum size of the main nodegroup. Must be larger than min_size.

__eks_main_min_size__ Defaults to 2, minimum size of the main nodegroup. Set to 0 to disablet his nodegroup.

__eks_main_instance_types__ List of instance types, defaults to  ["t3.large"]. Set according to clients needs.

__eks_spot_min_size__ Defaults to 1, minimum size of the nodegroup. Set to 0 to disablet his nodegroup.

__eks_spot_max_size__ Defaults to 3, maximum size of the nodegroup. Must be larger than min_size.

__eks_spot_instance_types__List of instance types, defaults to  ["t3.medium", "t3.large"]. Set according to clients needs.

__eks_db_min_size__ Defaults to 1, minimum size of the nodegroup. Set to 0 to disablet his nodegroup.

__eks_db_max_size__ Defaults to 3, maximum size of the nodegroup. Must be larger than min_size.

__eks_db_instance_types__List of instance types, defaults to  ["t3.medium", "t3.large"]. Set according to clients needs.

__eks_mon_min_size__ Defaults to 1, minimum size of the nodegroup. Set to 0 to disablet his nodegroup.

__eks_mon_max_size__ Defaults to 3, maximum size of the nodegroup. Must be larger than min_size.

__eks_mon_instance_types__ List of instance types, defaults to  ["t3.large"]. Set according to clients needs.

__eks_mon_single_subnet__ Defaults to true, if set to false then monitoring nodegroup nodes will not be forced into one subnet.

__cluster_enabled_log_types__ Defaults to ["api", "authenticator"], to disable logging set to [].

__crossplane_enable__ Defaults to true, Creates needed IRSA, Conigmap and SSM parameters for crossplane. If modules/k8s/crossplane is used then set to true.

### SSM parameters ###
**Only if var.crossplane_enable is true'**
```

"/entigo-infralib/${local.hname}/eks/account"
"/entigo-infralib/${local.hname}/eks/region"
"/entigo-infralib/${local.hname}/eks/oidc"

```


### Example code ###

```
    modules:
      - name: eks
        source: aws/eks
        inputs:
          vpc_prefix: "ep-network-vpc"
          eks_main_min_size: 3
          eks_main_max_size: 6
          eks_spot_max_size: 0
          eks_db_max_size: 0
          cluster_enabled_log_types: |
            []

```

### Limitations ###
If you want multiple eks clusters then they have to be under different workspaces. The module name has to be "eks" - otherwise the terraform provider for kubernets and helm will be misconfigured.
