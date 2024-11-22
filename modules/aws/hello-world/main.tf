resource "aws_ssm_parameter" "hello_world" {
  name  = "/entigo-infralib/${var.prefix}/hello_world"
  type  = "String"
  value = "Hello, ${var.prefix}!"
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
  }
}
