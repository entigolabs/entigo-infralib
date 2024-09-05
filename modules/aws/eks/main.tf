locals {
  auth_roles = [
    {
      rolearn  = replace(element(tolist(data.aws_iam_roles.aws-admin-roles.arns), 0), "//aws-reserved.*/AWSReservedSSO/", "/AWSReservedSSO")
      username = "aws-admin"
      groups   = ["system:masters"]
    },
    # {
    #   rolearn  = replace(element(tolist(data.aws_iam_roles.aws-admin-roles.arns), 0), "//aws-reserved.*", "${local.hname}-karpenter-node-role")
    #   username = "system:node:{{EC2PrivateDNSName}}"
    #   groups   = ["system:nodes", "system:bootstrappers"]
    # },
  ]

  eks_managed_node_groups_all = {
    main = {
      min_size        = var.eks_main_min_size
      desired_size    = var.eks_main_desired_size != 0 ? var.eks_main_desired_size : var.eks_main_min_size
      max_size        = var.eks_main_max_size
      instance_types  = var.eks_main_instance_types
      capacity_type   = "ON_DEMAND"
      release_version = var.eks_cluster_version

      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        Workspace = terraform.workspace
      }
      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_main_volume_size
            volume_iops           = var.eks_main_volume_iops
            volume_type           = var.eks_main_volume_type
            delete_on_termination = true
          }
        }
      }
    },
    mainarm = {
      min_size        = var.eks_mainarm_min_size
      desired_size    = var.eks_mainarm_desired_size != 0 ? var.eks_mainarm_desired_size : var.eks_mainarm_min_size
      max_size        = var.eks_mainarm_max_size
      instance_types  = var.eks_mainarm_instance_types
      capacity_type   = "ON_DEMAND"
      release_version = var.eks_cluster_version
      ami_type        = "AL2_ARM_64"
      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        Workspace = terraform.workspace
      }
      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_mainarm_volume_size
            volume_iops           = var.eks_mainarm_volume_iops
            volume_type           = var.eks_mainarm_volume_type
            delete_on_termination = true
          }
        }
      }
    },
    spot = {
      min_size        = var.eks_spot_min_size
      desired_size    = var.eks_spot_desired_size != 0 ? var.eks_spot_desired_size : var.eks_spot_min_size
      max_size        = var.eks_spot_max_size
      instance_types  = var.eks_spot_instance_types
      capacity_type   = "SPOT"
      release_version = var.eks_cluster_version

      taints = [
        {
          key    = "spot"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]
      labels = {
        spot = "true"
      }
      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        Workspace = terraform.workspace
      }
      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_spot_volume_size
            volume_iops           = var.eks_spot_volume_iops
            volume_type           = var.eks_spot_volume_type
            delete_on_termination = true
          }
        }
      }
    },
    mon = {
      min_size        = var.eks_mon_min_size
      desired_size    = var.eks_mon_desired_size != 0 ? var.eks_mon_desired_size : var.eks_mon_min_size
      max_size        = var.eks_mon_max_size
      instance_types  = var.eks_mon_instance_types
      subnet_ids      = var.eks_mon_single_subnet ? [var.private_subnets[0]] : var.private_subnets
      capacity_type   = "ON_DEMAND"
      release_version = var.eks_cluster_version
      taints = [
        {
          key    = "mon"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]
      labels = {
        mon = "true"
      }
      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        Workspace = terraform.workspace
      }

      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_mon_volume_size
            volume_iops           = var.eks_mon_volume_iops
            volume_type           = var.eks_mon_volume_type
            delete_on_termination = true
          }
        }
      }
    },
    tools = {
      min_size        = var.eks_tools_min_size
      desired_size    = var.eks_tools_desired_size != 0 ? var.eks_tools_desired_size : var.eks_tools_min_size
      max_size        = var.eks_tools_max_size
      instance_types  = var.eks_tools_instance_types
      subnet_ids      = var.eks_tools_single_subnet ? [var.private_subnets[0]] : var.private_subnets
      capacity_type   = "ON_DEMAND"
      release_version = var.eks_cluster_version
      taints = [
        {
          key    = "tools"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]
      labels = {
        tools = "true"
      }
      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        Workspace = terraform.workspace
      }

      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_tools_volume_size
            volume_iops           = var.eks_tools_volume_iops
            volume_type           = var.eks_tools_volume_type
            delete_on_termination = true
          }
        }
      }
    },
    db = {
      min_size        = var.eks_db_min_size
      desired_size    = var.eks_db_desired_size != 0 ? var.eks_db_desired_size : var.eks_db_min_size
      max_size        = var.eks_db_max_size
      instance_types  = var.eks_db_instance_types
      capacity_type   = "ON_DEMAND"
      release_version = var.eks_cluster_version
      taints = [
        {
          key    = "db"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]
      labels = {
        db = "true"
      }
      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        Workspace = terraform.workspace
      }

      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_db_volume_size
            volume_iops           = var.eks_db_volume_iops
            volume_type           = var.eks_db_volume_type
            delete_on_termination = true
          }
        }
      }
    }
  }

  # Need to keep role name_prefix length under 38. 
  eks_managed_node_groups_default = {
    for key, value in local.eks_managed_node_groups_all :
    "${substr(local.hname, 0, 21 - length(key) >= 0 ? 21 - length(key) : 0)}${length(key) < 21 ? "-" : ""}${substr(key, 0, 22)}" => value if key == "main" && var.eks_main_max_size > 0 || key == "mainarm" && var.eks_mainarm_max_size > 0 || key == "spot" && var.eks_spot_max_size > 0 || key == "mon" && var.eks_mon_max_size > 0 || key == "tools" && var.eks_tools_max_size > 0 || key == "db" && var.eks_db_max_size > 0
  }

  # Set desired_size to min_size if desired_size is 0 for extra node groups
  eks_managed_node_groups_extra = {
    for k, v in var.eks_managed_node_groups_extra :
    k => merge(
      v,
      {
        desired_size = lookup(v, "desired_size", 0) > 0 ? v.desired_size : lookup(v, "min_size", 1)
      }
    )
  }

  eks_managed_node_groups = merge(local.eks_managed_node_groups_default, local.eks_managed_node_groups_extra)

  extra_min_sizes     = { for node_group_name, node_group_config in var.eks_managed_node_groups_extra : "eks_${node_group_name}_min_size" => lookup(node_group_config, "min_size", 1) }
  extra_desired_sizes = { for node_group_name, node_group_config in var.eks_managed_node_groups_extra : "eks_${node_group_name}_desired_size" => lookup(node_group_config, "desired_size", 0) }

  // Contains desired sizes with values more than 0
  eks_desired_size_map = {
    for k, v in merge(
      {
        eks_main_desired_size    = var.eks_main_desired_size
        eks_mainarm_desired_size = var.eks_mainarm_desired_size
        eks_tools_desired_size   = var.eks_tools_desired_size
        eks_mon_desired_size     = var.eks_mon_desired_size
        eks_spot_desired_size    = var.eks_spot_desired_size
        eks_db_desired_size      = var.eks_db_desired_size
      },
      local.extra_desired_sizes
    ) : k => v if v > 0
  }

  // Contains min sizes for node pools that have desired size value more than 0
  eks_min_size_map = {
    for k, v in merge(
      {
        eks_main_min_size    = var.eks_main_min_size
        eks_mainarm_min_size = var.eks_mainarm_min_size
        eks_tools_min_size   = var.eks_tools_min_size
        eks_mon_min_size     = var.eks_mon_min_size
        eks_spot_min_size    = var.eks_spot_min_size
        eks_db_min_size      = var.eks_db_min_size
      },
      local.extra_min_sizes
    ) : k => v if contains(keys(local.eks_desired_size_map), replace(k, "min_size", "desired_size"))
  }

  temp_map_1 = {
    for k, v in local.eks_min_size_map : k => v if local.eks_desired_size_map[replace(k, "min_size", "desired_size")] >= v
  }

  temp_map_2 = {
    for k, v in local.eks_desired_size_map : k => v if local.eks_min_size_map[replace(k, "desired_size", "min_size")] <= v
  }

  // Contains min_size and desired_size for node groups that have desired_size >= min_size
  eks_min_and_desired_size_map = merge(local.temp_map_1, local.temp_map_2)
}

resource "aws_ec2_tag" "privatesubnets" {
  for_each    = toset(var.private_subnets)
  resource_id = each.key
  key         = "kubernetes.io/cluster/${local.hname}"
  value       = "shared"
}

resource "aws_ec2_tag" "publicsubnets" {
  for_each    = toset(var.public_subnets)
  resource_id = each.key
  key         = "kubernetes.io/cluster/${local.hname}"
  value       = "shared"
}

module "ebs_csi_irsa_role" {
  source                = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version               = "5.39.0"
  role_name             = "${local.hname}-ebs-csi"
  attach_ebs_csi_policy = true
  oidc_providers = {
    ex = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:ebs-csi-controller-sa"]
    }
  }
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

module "vpc_cni_irsa_role" {
  source                = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version               = "5.39.0"
  role_name_prefix      = "VPC-CNI-IRSA"
  attach_vpc_cni_policy = true
  vpc_cni_enable_ipv4   = true
  vpc_cni_enable_ipv6   = true

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:aws-node"]
    }
  }
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

#https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "19.21.0"

  cluster_name                    = local.hname
  cluster_version                 = var.eks_cluster_version
  cluster_endpoint_private_access = true
  cluster_endpoint_public_access  = var.eks_cluster_public
  cluster_enabled_log_types       = var.cluster_enabled_log_types
  enable_irsa                     = true

  cluster_addons = {
    coredns = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version               = "v1.10.1-eksbuild.7"
      configuration_values = jsonencode({
        tolerations : [
          {
            key : "tools",
            operator : "Equal",
            value : "true",
            effect : "NoSchedule"
          }
        ],
        affinity : {
          nodeAffinity : {
            preferredDuringSchedulingIgnoredDuringExecution : [
              {
                preference : {
                  matchExpressions : [
                    {
                      "key" : "tools",
                      "operator" : "In",
                      "values" : [
                        "true"
                      ]
                    }
                  ]
                },
                "weight" : 5
              }
            ]
          }
        }
      })
    }
    kube-proxy = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version               = "v1.28.8-eksbuild.2"
    }
    vpc-cni = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version               = "v1.18.0-eksbuild.1"
      most_recent                 = true
      before_compute              = true
      service_account_role_arn    = module.vpc_cni_irsa_role.iam_role_arn

      configuration_values = jsonencode({
        env = {
          ENABLE_PREFIX_DELEGATION = "true"
          WARM_PREFIX_TARGET       = "1"
        }
      })
    }
    aws-ebs-csi-driver = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version               = "v1.30.0-eksbuild.1"
      #configuration_values     = "{\"controller\":{\"extraVolumeTags\": {\"map-migrated\": \"migXXXXX\"}}}"
      service_account_role_arn = module.ebs_csi_irsa_role.iam_role_arn
      configuration_values = jsonencode({
        controller : {
          tolerations : [
            {
              key : "tools",
              operator : "Equal",
              value : "true",
              effect : "NoSchedule"
            }
          ],
          affinity : {
            nodeAffinity : {
              preferredDuringSchedulingIgnoredDuringExecution : [
                {
                  preference : {
                    matchExpressions : [
                      {
                        "key" : "eks.amazonaws.com/compute-type",
                        "operator" : "NotIn",
                        "values" : [
                          "fargate"
                        ]
                      }
                    ]
                  },
                  "weight" : 1
                },
                {
                  preference : {
                    matchExpressions : [
                      {
                        "key" : "tools",
                        "operator" : "In",
                        "values" : [
                          "true"
                        ]
                      }
                    ]
                  },
                  "weight" : 5
                }
              ]
            }
          }
        }
      })
    }
  }

  vpc_id     = var.vpc_id
  subnet_ids = var.private_subnets

  cluster_security_group_additional_rules = {
    egress_nodes_ephemeral_ports_tcp = {
      description                = "To node 1025-65535"
      protocol                   = "tcp"
      from_port                  = 1025
      to_port                    = 65535
      type                       = "egress"
      source_node_security_group = true
    }
    ingress_private = {
      description = "From self private"
      protocol    = "tcp"
      from_port   = 443
      to_port     = 443
      type        = "ingress"
      cidr_blocks = var.eks_api_access_cidrs
    }
  }

  node_security_group_additional_rules = {
    sidecar_injection_for_istio = {
      type                          = "ingress"
      protocol                      = "tcp"
      from_port                     = 15017
      to_port                       = 15017
      source_cluster_security_group = true
      description                   = "Allow istio to inject sidecars"
    }
    ingress_from_control_plane = {
      type                          = "ingress"
      protocol                      = "tcp"
      from_port                     = 8080
      to_port                       = 8080
      source_cluster_security_group = true
      description                   = "Allow http from control plane"
    }
    ingress_self_all = {
      description = "Node to node all ports/protocols"
      protocol    = "-1"
      from_port   = 0
      to_port     = 0
      type        = "ingress"
      self        = true
    }
    egress_all = {
      description = "Node all egress"
      protocol    = "-1"
      from_port   = 0
      to_port     = 0
      type        = "egress"
      cidr_blocks = ["0.0.0.0/0"]
    }

    ingress_allow_nodeport = {
      description = "Allow NodePort"
      protocol    = "-1"
      from_port   = 0
      to_port     = 0
      type        = "ingress"
      cidr_blocks = var.eks_nodeport_access_cidrs
    }

  }

  #https://github.com/terraform-aws-modules/terraform-aws-eks/issues/1986
  node_security_group_tags = {
    "kubernetes.io/cluster/${local.hname}" = null
    "karpenter.sh/discovery" = local.hname
  }

  cluster_encryption_config = []

  eks_managed_node_group_defaults = {
    iam_role_additional_policies = {
      AmazonSSMManagedInstanceCore = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
    }
    iam_role_attach_cni_policy = false
  }

  eks_managed_node_groups = local.eks_managed_node_groups

  # aws-auth configmap
  manage_aws_auth_configmap = true
  create_aws_auth_configmap = false
  aws_auth_roles            = local.auth_roles

  aws_auth_users = [
  ]

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "cluster_name" {
  name  = "/entigo-infralib/${local.hname}/cluster_name"
  type  = "String"
  value = local.hname
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "account" {
  name  = "/entigo-infralib/${local.hname}/account"
  type  = "String"
  value = data.aws_caller_identity.current.account_id
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "region" {
  name  = "/entigo-infralib/${local.hname}/region"
  type  = "String"
  value = data.aws_region.current.name
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_oidc_provider" {
  name  = "/entigo-infralib/${local.hname}/oidc_provider"
  type  = "String"
  value = module.eks.oidc_provider
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_oidc_provider_arn" {
  name  = "/entigo-infralib/${local.hname}/oidc_provider_arn"
  type  = "String"
  value = module.eks.oidc_provider_arn
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

#resource "aws_eks_identity_provider_config" "aad" {
#  cluster_name = module.eks.cluster_name
#  oidc {
#    client_id                     = "9995b0f0-1d59-48a7-8feb-7a58f6879833"
#    identity_provider_config_name = "AAD"
#    issuer_url                    = "https://sts.windows.net/cee3f45d-55bb-4dd1-b79b-111c9738f9df/"
#    username_claim                = "upn"
#    groups_claim                  = "groups"
#  }
#}

resource "null_resource" "update_desired_size" {
  count      = length(local.eks_desired_size_map) > 0 ? 1 : 0
  depends_on = [module.eks]

  triggers = {
    eks_desired_size_map = jsonencode([
      for key in sort(keys(local.eks_desired_size_map)) : {
        key   = key
        value = local.eks_desired_size_map[key]
      }
    ])
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    environment = local.eks_min_and_desired_size_map

    command = <<-EOT

      # Get list of node groups
      nodegroups=$(aws eks list-nodegroups --cluster-name ${module.eks.cluster_name} --query "nodegroups" --output text)
      
      # Loop through each node group
      for nodegroup in $nodegroups; do
        echo ""
        echo "Nodegroup: $nodegroup"

        # Get the short name of the node group (Example: main, mainarm, tools, mon, spot, db)
        node_group_short_name=$(echo "$nodegroup" | awk -F'-' '{print $(NF-1)}')
        echo "Node group short name: $node_group_short_name"

        # Get desired size variable name (Example: eks_main_desired_size)
        desired_size_variable_name="eks_$${node_group_short_name}_desired_size"
        echo "desired_size_variable_name: $desired_size_variable_name"

        # If desired_size_variable_name is not set as an environment variable (Does not exist in eks_min_and_desired_size_map), skip this node group
        if [ -z "$${!desired_size_variable_name}" ]; then
          echo "Skipping node group $nodegroup because desired_size variable is not set"
          continue
        fi

        # Get new desired size value from environment variable
        new_desired_size=$${!desired_size_variable_name}

        # Convert new desired size value to an integer
        new_desired_size=$(printf "%d" "$new_desired_size")
        echo "New desired size: $new_desired_size"
        
        # Get the current desired size of the node group
        current_desired_size=$(aws eks describe-nodegroup --cluster-name ${module.eks.cluster_name} --nodegroup-name $nodegroup --query "nodegroup.scalingConfig.desiredSize" --output text)

        # Convert current desired size value to an integer
        current_desired_size=$(printf "%d" "$current_desired_size")
        echo "Current desired size: $current_desired_size"

        if [ $current_desired_size -eq $new_desired_size ]; then
           echo "Node group $nodegroup already at desired size: $new_desired_size". No update needed.
           continue
        fi

        # Get min size variable name (Example: eks_main_min_size)
        min_size_variable_name="eks_$${node_group_short_name}_min_size"
        echo "min_size_variable_name: $min_size_variable_name"

        # Get new min size value from environment variable
        new_min_size=$${!min_size_variable_name}

        # Convert new min size value to an integer
        new_min_size=$(printf "%d" "$new_min_size")
        echo "New min size: $new_min_size"

        # Get the current min size of the node group
        current_min_size=$(aws eks describe-nodegroup --cluster-name ${module.eks.cluster_name} --nodegroup-name $nodegroup --query "nodegroup.scalingConfig.minSize" --output text)

        # Convert current min size value to an integer
        current_min_size=$(printf "%d" "$current_min_size")
        echo "Current min size: $current_min_size"

        # Check if node group is in ACTIVE state, if not then sleep for 5 seconds and check again
        while [ $(aws eks describe-nodegroup --cluster-name ${module.eks.cluster_name} --nodegroup-name $nodegroup --query "nodegroup.status" --output text) != "ACTIVE" ]; do
          sleep 5
        done

        # Update node group desired size
        aws eks update-nodegroup-config --cluster-name ${module.eks.cluster_name} --nodegroup-name $nodegroup --scaling-config desiredSize=$new_desired_size
        echo "Updated node group $nodegroup to new desired size: $new_desired_size"

        # Check if node group is in ACTIVE state, if not then sleep for 5 seconds and check again
        while [ $(aws eks describe-nodegroup --cluster-name ${module.eks.cluster_name} --nodegroup-name $nodegroup --query "nodegroup.status" --output text) != "ACTIVE" ]; do
          sleep 5
        done

      done

    EOT
  }
}