output "local_hname" {
  value = local.hname
}

output "eks_main_min_size" {
  value = var.eks_main_min_size
}

output "eks_mainarm_min_size" {
  value = var.eks_mainarm_min_size
}

output "eks_tools_min_size" {
  value = var.eks_tools_min_size
}

output "eks_mon_min_size" {
  value = var.eks_mon_min_size
}

output "eks_spot_min_size" {
  value = var.eks_spot_min_size
}

output "eks_db_min_size" {
  value = var.eks_db_min_size
}

output "eks_extra_min_sizes" {
  value = var.eks_extra_min_sizes
}