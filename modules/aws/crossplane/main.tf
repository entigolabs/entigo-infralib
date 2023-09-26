resource "aws_iam_role" "crossplane" {
  name = "crossplane-${local.hname}"

  assume_role_policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "${data.aws_ssm_parameter.eks_oidc_provider_arn.value}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "${data.aws_ssm_parameter.eks_oidc_provider.value}:aud": "sts.amazonaws.com",
                    "${data.aws_ssm_parameter.eks_oidc_provider.value}:sub": "system:serviceaccount:crossplane-system:aws-crossplane"
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
      awsAccount  = data.aws_ssm_parameter.account.value
      awsRegion   = data.aws_ssm_parameter.region.value
      clusterOIDC = data.aws_ssm_parameter.eks_oidc_provider.value
    })
  ]
}
