terraform {
  backend "s3" {}
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">4"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~>2.0"
    }
  }
}

provider "aws" {
}

provider "kubernetes" {
  host                   = module.test_eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.test_eks.cluster_certificate_authority_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.test_eks.cluster_name]
  }
}


# Test with default values
module "test_eks" {
  source = "../"

  prefix                        = var.prefix
  vpc_prefix                    = var.vpc_prefix
  eks_cluster_version           = var.eks_cluster_version
  iam_admin_role                = var.iam_admin_role
  eks_cluster_public            = var.eks_cluster_public
  eks_main_min_size             = var.eks_main_min_size
  eks_main_max_size             = var.eks_main_max_size
  eks_main_instance_types       = var.eks_main_instance_types
  eks_spot_min_size             = var.eks_spot_min_size
  eks_spot_max_size             = var.eks_spot_max_size
  eks_spot_instance_types       = var.eks_spot_instance_types
  eks_db_min_size               = var.eks_db_min_size
  eks_db_max_size               = var.eks_db_max_size
  eks_db_instance_types         = var.eks_db_instance_types
  eks_monitoring_min_size       = var.eks_monitoring_min_size
  eks_monitoring_max_size       = var.eks_monitoring_max_size
  eks_monitoring_instance_types = var.eks_monitoring_instance_types
  eks_monitoring_single_subnet  = var.eks_monitoring_single_subnet
  cluster_enabled_log_types     = var.cluster_enabled_log_types
}

output "cluster_name" {
  value = module.test_eks.cluster_name
}
