locals {
  auth_roles = [
    {
      rolearn  = replace(element(tolist(data.aws_iam_roles.aws-admin-roles.arns), 0), "//aws-reserved.*/AWSReservedSSO/", "/AWSReservedSSO")
      username = "aws-admin"
      groups   = ["system:masters"]
    }
  ]

  eks_managed_node_groups_all = {
    main = {
      min_size        = var.eks_main_min_size
      desired_size    = var.eks_main_min_size
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
            volume_size           = 100
            volume_iops           = 3000
            volume_type           = "gp3"
            delete_on_termination = true
          }
        }
      }
    },
    mainarm = {
      min_size        = var.eks_mainarm_min_size
      desired_size    = var.eks_mainarm_min_size
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
            volume_size           = 100
            volume_iops           = 3000
            volume_type           = "gp3"
            delete_on_termination = true
          }
        }
      }
    },
    spot = {
      min_size        = var.eks_spot_min_size
      desired_size    = var.eks_spot_min_size
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
            volume_size           = 100
            volume_iops           = 3000
            volume_type           = "gp3"
            delete_on_termination = true
          }
        }
      }
    },
    mon = {
      min_size        = var.eks_mon_min_size
      desired_size    = var.eks_mon_min_size
      max_size        = var.eks_mon_max_size
      instance_types  = var.eks_mon_instance_types
      subnet_ids      = var.eks_mon_single_subnet ? [split(",", data.aws_ssm_parameter.private_subnets.value)[0]] : split(",", data.aws_ssm_parameter.private_subnets.value)
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
            volume_size           = 50
            volume_iops           = 3000
            volume_type           = "gp3"
            delete_on_termination = true
          }
        }
      }
    },
    tools = {
      min_size        = var.eks_tools_min_size
      desired_size    = var.eks_tools_min_size
      max_size        = var.eks_tools_max_size
      instance_types  = var.eks_tools_instance_types
      subnet_ids      = var.eks_tools_single_subnet ? [split(",", data.aws_ssm_parameter.private_subnets.value)[0]] : split(",", data.aws_ssm_parameter.private_subnets.value)
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
            volume_size           = 50
            volume_iops           = 3000
            volume_type           = "gp3"
            delete_on_termination = true
          }
        }
      }
    },
    db = {
      min_size        = var.eks_db_min_size
      desired_size    = var.eks_db_min_size
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
            volume_size           = 50
            volume_iops           = 3000
            volume_type           = "gp3"
            delete_on_termination = true
          }
        }
      }
    }
  }

  # Need to keep role name_prefix length under 38. 
  eks_managed_node_groups = {
    for key, value in local.eks_managed_node_groups_all :
    "${substr(local.hname, 0, 21 - length(key) >= 0 ? 21 - length(key) : 0)}${length(key) < 21 ? "-" : ""}${substr(key, 0, 22)}" => value if key == "main" && var.eks_main_max_size > 0 || key == "mainarm" && var.eks_mainarm_max_size > 0 || key == "spot" && var.eks_spot_max_size > 0 || key == "mon" && var.eks_mon_max_size > 0 || key == "tools" && var.eks_tools_max_size > 0 || key == "db" && var.eks_db_max_size > 0

  }

}

resource "aws_ec2_tag" "privatesubnets" {
  for_each    = toset(split(",", nonsensitive(data.aws_ssm_parameter.private_subnets.value)))
  resource_id = each.key
  key         = "kubernetes.io/cluster/${local.hname}"
  value       = "shared"
}

resource "aws_ec2_tag" "publicsubnets" {
  for_each    = toset(split(",", nonsensitive(data.aws_ssm_parameter.public_subnets.value)))
  resource_id = each.key
  key         = "kubernetes.io/cluster/${local.hname}"
  value       = "shared"
}

module "ebs_csi_irsa_role" {
  source                = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version               = "5.27.0"
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
  version               = "5.27.0"
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
  version = "19.15.3"

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
      addon_version = "v1.9.3-eksbuild.3"
      configuration_values = jsonencode({
          tolerations: [
            {
            key: "tools",
            operator: "Equal",
            value: "true",
            effect: "NoSchedule"
            }
          ],
          affinity: {
              nodeAffinity: {
                preferredDuringSchedulingIgnoredDuringExecution: [
                  {
                    preference: {
                      matchExpressions: [
                       {
                          "key": "tools",
                          "operator": "In",
                          "values": [
                            "true"
                          ]
                        }
                      ]
                    },
                    "weight": 5
                  }
                ]
              }
          }
      })
    }
    kube-proxy = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version = "v1.26.2-eksbuild.1"
    }
    vpc-cni = {
      resolve_conflicts_on_update = "OVERWRITE"
      resolve_conflicts_on_create = "OVERWRITE"
      addon_version = "v1.15.0-eksbuild.2"
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
      addon_version = "v1.23.0-eksbuild.1"
      #configuration_values     = "{\"controller\":{\"extraVolumeTags\": {\"map-migrated\": \"migXXXXX\"}}}"
      service_account_role_arn = module.ebs_csi_irsa_role.iam_role_arn
      configuration_values = jsonencode({
        controller: {
          tolerations: [
            {
            key: "tools",
            operator: "Equal",
            value: "true",
            effect: "NoSchedule"
            }
          ],
          affinity: {
              nodeAffinity: {
                preferredDuringSchedulingIgnoredDuringExecution: [
                  {
                    preference: {
                      matchExpressions: [
                       {
                          "key": "eks.amazonaws.com/compute-type",
                          "operator": "NotIn",
                          "values": [
                            "fargate"
                          ]
                        }
                      ]
                    },
                    "weight": 1
                  },
                  {
                    preference: {
                      matchExpressions: [
                       {
                          "key": "tools",
                          "operator": "In",
                          "values": [
                            "true"
                          ]
                        }
                      ]
                    },
                    "weight": 5
                  }
                ]
              }
          }
        }
      })
    }
  }

  vpc_id     = data.aws_ssm_parameter.vpc_id.value
  subnet_ids = split(",", data.aws_ssm_parameter.private_subnets.value)

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
      cidr_blocks = split(",", data.aws_ssm_parameter.private_subnet_cidrs.value)
    }
    #ingress_intranet = {
    #  description = "From self intranet"
    #  protocol    = "tcp"
    #  from_port   = 443
    #  to_port     = 443
    #  type        = "ingress"
    #  cidr_blocks = split(",", data.aws_ssm_parameter.intra_subnet_cidrs.value)
    #}
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
  }

  #https://github.com/terraform-aws-modules/terraform-aws-eks/issues/1986
  node_security_group_tags = {
    "kubernetes.io/cluster/${local.hname}" = null
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

resource "aws_ssm_parameter" "account" {
  name  = "/entigo-infralib/${local.hname}/eks/account"
  type  = "String"
  value = data.aws_caller_identity.current.account_id
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "region" {
  name  = "/entigo-infralib/${local.hname}/eks/region"
  type  = "String"
  value = data.aws_region.current.name
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_oidc" {
  name  = "/entigo-infralib/${local.hname}/eks/oidc"
  type  = "String"
  value = module.eks.oidc_provider
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_oidc_provider" {
  name  = "/entigo-infralib/${local.hname}/eks/oidc_provider"
  type  = "String"
  value = module.eks.oidc_provider
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_oidc_provider_arn" {
  name  = "/entigo-infralib/${local.hname}/eks/oidc_provider_arn"
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

