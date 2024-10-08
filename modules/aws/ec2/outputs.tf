
resource "aws_ssm_parameter" "public_ip" {
  count = var.eip ? 1 : 0
  name  = "/entigo-infralib/${var.prefix}/public_ip"
  type  = "String"
  value = aws_eip.ec2[0].public_ip
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}

output "private_dns" {
  value = aws_instance.ec2.private_dns
}
