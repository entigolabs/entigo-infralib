resource "aws_iam_role" "crossplane" {
  count = (var.crossplane_enable) ? 0 : 1
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
  count = (var.crossplane_enable) ? 0 : 1
  role       = aws_iam_role.crossplane[0].name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
}

resource "kubernetes_namespace" "crossplane-system" {
  count = (var.crossplane_enable) ? 0 : 1
  metadata {
    name = "corssplane-system"
  }
}

resource "kubernetes_service_account" "aws-crossplane" {
  count = (var.crossplane_enable) ? 0 : 1
  metadata {
    name = "aws-crossplane"
    namespace = kubernetes_namespace.crossplane-system.name
    annotations = {
     "eks.amazonaws.com/role-arn" = aws_iam_role.crossplane[0].arn
    }
  }
}
