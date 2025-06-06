locals {
  #Cant do yet...
  #https://github.com/terraform-aws-modules/terraform-aws-eks/issues/3036 
  #cluster_encryption_config = var.cluster_encryption_kms_key_arn != null ? [{
  #  resources        = ["secrets"]
  #  provider_key_arn = var.cluster_encryption_kms_key_arn
  #}] : []
  cluster_encryption_config = {}
  
  iam_role_additional_policies = zipmap(compact(var.iam_role_additional_policies), compact(var.iam_role_additional_policies))

  eks_managed_node_groups_all = {
    main = {
      min_size        = var.eks_main_min_size
      desired_size    = var.eks_main_desired_size != 0 ? var.eks_main_desired_size : var.eks_main_min_size
      max_size        = var.eks_main_max_size
      instance_types  = var.eks_main_instance_types
      subnet_ids      = length(var.eks_main_subnets) == 0 ? var.private_subnets : var.eks_main_subnets
      capacity_type   = var.eks_main_capacity_type
      key_name         = var.node_ssh_key_pair_name
      release_version = var.eks_cluster_version
      ami_type        = var.eks_main_ami_type
      labels = {
        main = "true"
      }
      autoscaling_group_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        created-by = "entigo-infralib"
      }
      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        created-by = "entigo-infralib"
      }
      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_main_volume_size
            volume_iops           = var.eks_main_volume_iops
            volume_type           = var.eks_main_volume_type
            encrypted             = var.node_encryption_kms_key_arn != "" ? true : false
            kms_key_id            = var.node_encryption_kms_key_arn != "" ? var.node_encryption_kms_key_arn : null
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
      subnet_ids      = length(var.eks_mon_subnets) == 0 ? var.private_subnets : var.eks_mon_subnets
      capacity_type   =  var.eks_mon_capacity_type
      key_name         = var.node_ssh_key_pair_name
      release_version = var.eks_cluster_version
      ami_type        = var.eks_mon_ami_type
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
      autoscaling_group_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        created-by = "entigo-infralib"
      }
      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        created-by = "entigo-infralib"
      }

      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_mon_volume_size
            volume_iops           = var.eks_mon_volume_iops
            volume_type           = var.eks_mon_volume_type
            encrypted             = var.node_encryption_kms_key_arn != "" ? true : false
            kms_key_id            = var.node_encryption_kms_key_arn != "" ? var.node_encryption_kms_key_arn : null
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
      subnet_ids      = length(var.eks_tools_subnets) == 0 ? var.private_subnets : var.eks_tools_subnets
      capacity_type   = var.eks_tools_capacity_type
      key_name         = var.node_ssh_key_pair_name
      release_version = var.eks_cluster_version
      ami_type        = var.eks_tools_ami_type
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
      autoscaling_group_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        created-by = "entigo-infralib"
      }
      launch_template_tags = {
        Terraform = "true"
        Prefix    = var.prefix
        created-by = "entigo-infralib"
      }

      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = var.eks_tools_volume_size
            volume_iops           = var.eks_tools_volume_iops
            volume_type           = var.eks_tools_volume_type
            encrypted             = var.node_encryption_kms_key_arn != "" ? true : false
            kms_key_id            = var.node_encryption_kms_key_arn != "" ? var.node_encryption_kms_key_arn : null
            delete_on_termination = true
          }
        }
      }
    }
  }

  # Need to keep role name_prefix length under 38. 
  eks_managed_node_groups_default = {
    for key, value in local.eks_managed_node_groups_all :
    "${substr(var.prefix, 0, 21 - length(key) >= 0 ? 21 - length(key) : 0)}${length(key) < 21 ? "-" : ""}${substr(key, 0, 22)}" => value if key == "main" && var.eks_main_max_size > 0 || key == "mon" && var.eks_mon_max_size > 0 || key == "tools" && var.eks_tools_max_size > 0
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
        eks_tools_desired_size   = var.eks_tools_desired_size
        eks_mon_desired_size     = var.eks_mon_desired_size
      },
      local.extra_desired_sizes
    ) : k => v if v > 0
  }

  // Contains min sizes for node pools that have desired size value more than 0
  eks_min_size_map = {
    for k, v in merge(
      {
        eks_main_min_size    = var.eks_main_min_size
        eks_tools_min_size   = var.eks_tools_min_size
        eks_mon_min_size     = var.eks_mon_min_size
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
  key         = "kubernetes.io/cluster/${var.prefix}"
  value       = "shared"
}

resource "aws_ec2_tag" "publicsubnets" {
  for_each    = toset(var.public_subnets)
  resource_id = each.key
  key         = "kubernetes.io/cluster/${var.prefix}"
  value       = "shared"
}

module "ebs_csi_irsa_role" {
  source                = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version               = "5.58.0"
  role_name             = "${var.prefix}-ebs-csi"
  attach_ebs_csi_policy = true
  ebs_csi_kms_cmk_ids = var.node_encryption_kms_key_arn != "" ? [var.node_encryption_kms_key_arn] : []
  oidc_providers = {
    ex = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:ebs-csi-controller-sa"]
    }
  }
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    created-by = "entigo-infralib"
  }
}

module "vpc_cni_irsa_role" {
  source                = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version               = "5.58.0"
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
    created-by = "entigo-infralib"
  }
}

#https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "20.36.0"

  cluster_name                    = var.prefix
  cluster_version                 = var.eks_cluster_version
  cluster_endpoint_private_access = true
  cluster_endpoint_public_access  = var.eks_cluster_public
  cluster_enabled_log_types       = var.cluster_enabled_log_types
  cloudwatch_log_group_kms_key_id = var.cloudwatch_log_group_kms_key_id != "" ? var.cloudwatch_log_group_kms_key_id : null
  
  cluster_identity_providers = var.cluster_identity_providers
  
  create_kms_key = false
  cluster_encryption_config = local.cluster_encryption_config

  create_iam_role = var.cluster_iam_role_arn != null ? false : true
  iam_role_arn = var.cluster_iam_role_arn

  enable_irsa                     = true

  bootstrap_self_managed_addons = var.bootstrap_self_managed_addons

  cluster_addons = {
    coredns = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version               = var.coredns_addon_version
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
      addon_version               = var.kube_proxy_addon_version
    }
    vpc-cni = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version               = var.vpc_cni_addon_version
      most_recent                 = true
      before_compute              = true
      service_account_role_arn    = module.vpc_cni_irsa_role.iam_role_arn

      configuration_values = jsonencode({
        env = {
          ENABLE_PREFIX_DELEGATION = var.enable_vpc_cni_prefix_delegation
          WARM_PREFIX_TARGET       = "1"
        }
      })
    }
    aws-ebs-csi-driver = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version               = var.ebs_csi_addon_version
      #configuration_values     = "{\"controller\":{\"extraVolumeTags\": {\"map-migrated\": \"migXXXXX\"}}}"
      service_account_role_arn = module.ebs_csi_irsa_role.iam_role_arn
      configuration_values = jsonencode({
        controller : {
          volumeModificationFeature: {
                enabled: true
          },
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
    "kubernetes.io/cluster/${var.prefix}" = null
    "karpenter.sh/discovery" = var.prefix
  }

  eks_managed_node_group_defaults = {
    iam_role_additional_policies = merge({ AmazonSSMManagedInstanceCore = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore" },
                                          local.iam_role_additional_policies)
    iam_role_attach_cni_policy = false
  }

  eks_managed_node_groups = local.eks_managed_node_groups

  # EKS access entries
  authentication_mode = var.authentication_mode
  enable_cluster_creator_admin_permissions = var.enable_cluster_creator_admin_permissions

  access_entries = merge({},
    var.iam_admin_role != "" ? {
      aws-admin = {
        principal_arn = element(tolist(data.aws_iam_roles.aws-admin-roles[0].arns), 0)
        user_name = "aws-admin"
        policy_associations = {
          aws-admin = {
            policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
            access_scope = {
              type = "cluster"
            }
          }
        }
      }
    } : {},
    var.aws_auth_user != ""? {
      aws-auth-user = {
        principal_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:user/${var.aws_auth_user}"
        user_name = var.aws_auth_user
        policy_associations = {
          aws-auth-user = {
            policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
            access_scope = {
              type = "cluster"
            }
          }
        }
      }
    } : {},
    var.additional_access_entries
  )

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    created-by = "entigo-infralib"
  }
}


#resource "aws_eks_identity_provider_config" "aad" {
#  cluster_name = module.eks.cluster_name
#  oidc {
#    client_id                     = "..."
#    identity_provider_config_name = "AAD"
#    issuer_url                    = "https://sts.windows.net/.../"
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

        # Get the short name of the node group (Example: main, tools, mon)
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
