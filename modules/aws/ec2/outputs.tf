
resource "aws_ssm_parameter" "public_ip" {
  count = var.eip ? 1 : 0
  name  = "/entigo-infralib/${local.hname}/public_ip"
  type  = "String"
  value = aws_eip.ec2[0].public_ip
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

output "private_dns" {
  value = aws_instance.ec2.private_dns
}
