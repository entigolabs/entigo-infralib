output "arn" {
  value = aws_instance.ec2.arn
}

output "id" {
  value = aws_instance.ec2.id
}

output "private_dns" {
  value = aws_instance.ec2.private_dns
}

output "public_ip" {
  value = var.eip ? aws_eip.ec2[0].public_ip : null
}
