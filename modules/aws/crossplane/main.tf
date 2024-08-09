resource "aws_iam_role" "crossplane" {
  name = "crossplane-${local.hname}"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "${var.eks_oidc_provider_arn}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "${var.eks_oidc_provider}:aud": "sts.amazonaws.com",
                    "${var.eks_oidc_provider}:sub": "system:serviceaccount:${var.kubernetes_namespace}:${var.kubernetes_service_account}"
                }
            }
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "crossplane-attach" {
  role       = aws_iam_role.crossplane.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

resource "aws_ssm_parameter" "iam_role" {
  name  = "/entigo-infralib/${local.hname}/iam_role"
  type  = "String"
  value = aws_iam_role.crossplane.arn
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}
