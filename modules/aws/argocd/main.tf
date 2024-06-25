resource "aws_iam_user_policy" "argocd" {
  name = "${local.hname}-argocd-aws"
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
  name = "${local.hname}-argocd-aws"
  path = "/"
}

resource "aws_iam_user_ssh_key" "argocd" {
  username   = aws_iam_user.argocd.name
  encoding   = "SSH"
  public_key = tls_private_key.argocd.public_key_openssh
}

locals {
   yqcomm = "yq -y -i '.version = \"${var.branch == "main" ? "0.1.0" : var.branch}\"' helm/modules/k8s/argocd/Chart.yaml"
}

data "external" "argocd" {
  program = ["bash", "-c", "rm -rf helm && git clone --depth 1 -b ${var.branch} ${var.repository} helm && ${local.yqcomm} && echo '{}'"]
}

locals {
  # This hash forces Terraform to redeploy if a new template file is added or changed, or values are updated
  namespace = var.namespace == "" ? "${local.hname}-aws" : var.namespace
  # chart_hash = sha1(join("", [for f in fileset("helm/modules/k8s/argocd", "**/*.yaml"): filesha1("helm/modules/k8s/argocd/${f}")]))
  values_template = templatefile("${path.module}/values.yaml", {
      hostname = var.hostname
      install_crd = var.install_crd
      workspace = terraform.workspace
      prefix = var.prefix
      argocd_apps_name = var.argocd_apps_name
      namespace = local.namespace
      ingress_group_name = var.ingress_group_name
      ingress_scheme = var.ingress_scheme
      sshPrivateKey = indent(10, tls_private_key.argocd.private_key_pem)
      repo = "ssh://${aws_iam_user_ssh_key.argocd.ssh_public_key_id}@git-codecommit.${data.aws_region.current.name}.amazonaws.com/v1/repos/${var.codecommit_name}"
  })
  values_hash = sha1(local.values_template)
  
}

resource "helm_release" "argocd" {
  name = var.name == "" ? "${local.hname}-argocd-aws" : var.name
  chart            = "helm/modules/k8s/argocd" 
  namespace        = local.namespace
  create_namespace = var.create_namespace
  values = [
    local.values_template
  ]
  set {
    name = "values-hash"
    value = local.values_hash
  }
  depends_on = [data.external.argocd]
}

resource "aws_ssm_parameter" "argocd_repo_url" {
  name  = "/entigo-infralib/${local.hname}/repo_url"
  type  = "String"
  value = "ssh://${aws_iam_user_ssh_key.argocd.ssh_public_key_id}@git-codecommit.${data.aws_region.current.name}.amazonaws.com/v1/repos/${var.codecommit_name}"
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "argocd_namespace" {
  name  = "/entigo-infralib/${local.hname}/namespace"
  type  = "String"
  value = local.namespace
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}
