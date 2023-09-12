resource "aws_iam_user_policy" "argocd" {
  name = "test"
  user = aws_iam_user.argocd.name
  policy = jsonencode({
          "Version": "2012-10-17",
          "Statement": [
              {
                  "Effect": "Allow",
                  "Action": [
                      "codecommit:BatchGet*",
                      "codecommit:BatchDescribe*",
                      "codecommit:Describe*",
                      "codecommit:EvaluatePullRequestApprovalRules",
                      "codecommit:Get*",
                      "codecommit:List*",
                      "codecommit:GitPull"
                  ],
                  "Resource": "*"
              }
          ]
      })
}

resource "tls_private_key" "argocd" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "aws_iam_user" "argocd" {
  name = "${local.hname}-argocd"
  path = "/"
}

resource "aws_iam_user_ssh_key" "argocd" {
  username   = aws_iam_user.argocd.name
  encoding   = "SSH"
  public_key = tls_private_key.argocd.public_key_openssh
}


resource "null_resource" "argocd" {
  provisioner "local-exec" {
    command = "git clone --depth 1 -b ${var.branch} ${var.repository} helm"
  }
  triggers = {
    always_run = timestamp()
  }
}

resource "helm_release" "argocd" {
  name = var.name == "" ? local.hname : var.name
  chart            = "helm/modules/k8s/argocd" 
  namespace        = var.namespace == "" ? local.hname : var.namespace
  create_namespace = var.create_namespace
  values = [
    templatefile("${path.module}/values.yaml", {
      hostname = var.hostname
      sshPrivateKey = tls_private_key.argocd.private_key_pem
      repo = "ssh://${aws_iam_user_ssh_key.argocd.ssh_public_key_id}@git-codecommit.${data.aws_region.current.name}.amazonaws.com/v1/repos/entigo-infralib-${data.aws_caller_identity.current.account_id}"
    })
  ]
  depends_on = [null_resource.argocd]
}
