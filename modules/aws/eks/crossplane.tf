resource "aws_iam_role" "crossplane" {
  count = var.crossplane_enable ? 1 : 0
  name = "crossplane-${local.hname}"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "${module.eks.oidc_provider_arn}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "${module.eks.oidc_provider}:aud": "sts.amazonaws.com",
                    "${module.eks.oidc_provider}:sub": "system:serviceaccount:crossplane-system:aws-crossplane"
                }
            }
        }
    ]
}
POLICY
}


resource "aws_iam_role_policy_attachment" "crossplane-attach" {
  count = var.crossplane_enable ? 1 : 0
  role       = aws_iam_role.crossplane[0].name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

resource "kubernetes_namespace" "crossplane-system" {
  count = var.crossplane_enable ? 1 : 0
  metadata {
    name = "crossplane-system"
  }
  depends_on = [
    module.eks
  ]
}

resource "kubernetes_service_account" "aws-crossplane" {
  count = var.crossplane_enable ? 1 : 0
  metadata {
    name = "aws-crossplane"
    namespace = kubernetes_namespace.crossplane-system[0].metadata[0].name
    annotations = {
     "eks.amazonaws.com/role-arn" = aws_iam_role.crossplane[0].arn
    }
  }
  depends_on = [
    module.eks
  ]
}

resource "kubernetes_config_map" "aws-crossplane" {
  count = var.crossplane_enable ? 1 : 0
  metadata {
    name = "aws-crossplane"
    namespace = kubernetes_namespace.crossplane-system[0].metadata[0].name
  }

  data = {
    awsAccount  = data.aws_caller_identity.current.account_id
    awsRegion   = data.aws_region.current.name
    clusterOIDC = module.eks.oidc_provider
  }
  depends_on = [
    module.eks
  ]
}

resource "aws_ssm_parameter" "account" {
  count = var.crossplane_enable ? 1 : 0
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
  count = var.crossplane_enable ? 1 : 0
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
  count = var.crossplane_enable ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/eks/oidc"
  type  = "String"
  value = module.eks.oidc_provider
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}
