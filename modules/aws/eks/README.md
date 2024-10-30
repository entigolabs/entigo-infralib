## Oppinionated module for eks creation ##


Oppinionated version of this https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest

__vpc_prefix__ look for subnets based on this SSM parameter prefix ("/entigo-infralib/${var.vpc_prefix}/vpc/vpc_id")


__vpc_id__ the VPC id where eks is installed.
  type = string
}

__private_subnets__ list(string) of the private subnet ID-s to use. Also nodegroups will be places in these subnets.

__public_subnets__ list(string) of the public subnet ID-s to use. This will only add the required tags for alb.

__eks_api_access_cidrs__ list(string) of the network ranges to allow api access. Usually private subents cidrs are allowed.

__eks_cluster_version__ SHOULD NOT BE CHANGED, but can override the EKS version to use.

__iam_admin_role__ Defaults to "AWSReservedSSO_AdministratorAccess_.*" but could be something else depending on SSO setup.

__eks_cluster_public__ Defaults to false, if set true then the EKS cluster endpoint will be on the public internet

__eks_main_min_size__ Defaults to 2, minimum size of the main nodegroup. Set to 0 to disablet his nodegroup.

__eks_main_max_size__ Defaults to 4, maximum size of the main nodegroup. Must be larger than min_size.

__eks_main_instance_types__ List of instance types, defaults to  ["t3.large"]. Set according to clients needs.

__eks_mainarm_min_size__ Defaults to 0, minimum size of the main nodegroup. Set to 0 to disablet his nodegroup.

__eks_mainarm_max_size__ Defaults to 0, maximum size of the main nodegroup. Must be larger than min_size.

__eks_mainarm_instance_types__ List of instance types, defaults to  ["t4g.large"]. Set according to clients needs.

__eks_spot_min_size__ Defaults to 0, minimum size of the nodegroup. Set to 0 to disablet his nodegroup.

__eks_spot_max_size__ Defaults to 0, maximum size of the nodegroup. Must be larger than min_size.

__eks_spot_instance_types__ List of instance types, defaults to  ["t3.medium", "t3.large"]. Set according to clients needs.

__eks_db_min_size__ Defaults to 0, minimum size of the nodegroup. Set to 0 to disablet his nodegroup.

__eks_db_max_size__ Defaults to 0, maximum size of the nodegroup. Must be larger than min_size.

__eks_db_instance_types__ List of instance types, defaults to  ["t3.medium", "t3.large"]. Set according to clients needs.

__eks_mon_min_size__ Defaults to 1, minimum size of the nodegroup. Set to 0 to disablet his nodegroup.

__eks_mon_max_size__ Defaults to 3, maximum size of the nodegroup. Must be larger than min_size.

__eks_mon_instance_types__ List of instance types, defaults to  ["t3.large"]. Set according to clients needs.

__eks_mon_single_subnet__ Defaults to true, if set to false then monitoring nodegroup nodes will not be forced into one subnet.

__eks_tools_min_size__ Defaults to 2, minimum size of the nodegroup. Set to 0 to disablet his nodegroup.

__eks_tools_max_size__ Defaults to 3, maximum size of the nodegroup. Must be larger than min_size.

__eks_tools_instance_types__ List of instance types, defaults to  ["t3.large"]. Set according to clients needs.

__eks_tools_single_subnet__ Defaults to false, if set to false then monitoring nodegroup nodes will not be forced into one subnet.

__cluster_enabled_log_types__ Defaults to ["api", "authenticator"], to disable logging set to [].

__eks_managed_node_groups_extra__ Defaults to {}, can add custom nodegroups or orverride defaults.

### SSM parameters ###
```
"/entigo-infralib/${var.prefix}/cluster_name"
"/entigo-infralib/${var.prefix}/account"
"/entigo-infralib/${var.prefix}/region"
"/entigo-infralib/${var.prefix}/oidc_provider_arn"
"/entigo-infralib/${var.prefix}/oidc_provider

```


### Example code ###

```
    modules:
      - name: eks
        source: aws/eks
        inputs:
          eks_main_min_size: 3
          eks_main_max_size: 6
          eks_spot_max_size: 0
          eks_db_max_size: 0
          cluster_enabled_log_types: |
            []

```

### Limitations ###
The module name has to be "eks" - otherwise the terraform provider for kubernets will be misconfigured.
