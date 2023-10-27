data "aws_eks_cluster" "eks" {
  name = var.eks_cluster_name
}

variable "eks_cluster_name" {
  type = string
}


provider "helm" {
  burst_limit = 300
  kubernetes {
    host                   = data.aws_eks_cluster.eks.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.eks.certificate_authority[0].data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      # This requires the awscli to be installed locally where Terraform is executed
      args = ["eks", "get-token", "--cluster-name", data.aws_eks_cluster.eks.name]
    }
  }
}
