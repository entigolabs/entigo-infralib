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
                    "${var.eks_oidc_provider}:sub": "system:serviceaccount:crossplane-system:aws-crossplane"
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


resource "helm_release" "crossplane-terraform" {
  name             = local.hname
  chart            = "${path. module}/helm" 
  namespace        = "crossplane-system"
  create_namespace = true
  values = [
    templatefile("${path.module}/values.yaml", {
      iamRole=aws_iam_role.crossplane.arn
      awsAccount  = var.eks_account
      awsRegion   = var.eks_region
      clusterOIDC = var.eks_oidc_provider
    })
  ]
}
