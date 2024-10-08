resource "aws_eip" "ec2" {
  count = var.eip ? 1 : 0
  instance = aws_instance.ec2.id
}

data "aws_ami" "ec2" {
  most_recent = true
  owners      = var.ami_owners
  filter {
    name   = "name"
    values = var.ami_names
  }
  filter {
      name   = "architecture"
      values = var.ami_architectures
  }
}

data "aws_subnet" "ec2" {
  id = var.subnet_id
}

resource "aws_security_group" "ec2" {
  name = var.prefix
  description = var.prefix
  vpc_id = data.aws_subnet.ec2.vpc_id
  tags = {
    "Name" = var.prefix
  }
}

resource "aws_security_group_rule" "egress" {
  for_each = var.security_group_egress
  type              = "egress"
  from_port         = each.value["from_port"]
  to_port           = each.value["to_port"]
  protocol          = each.value["protocol"]
  cidr_blocks       = each.value["cidr_blocks"]
  description = each.key
  security_group_id = aws_security_group.ec2.id
}

resource "aws_security_group_rule" "ingress" {
  for_each = var.security_group_ingress
  type              = "ingress"
  from_port         = each.value["from_port"]
  to_port           = each.value["to_port"]
  protocol          = each.value["protocol"]
  cidr_blocks       = each.value["cidr_blocks"]
  description = each.key
  security_group_id = aws_security_group.ec2.id
}

resource "aws_instance" "ec2" {
  ami           = data.aws_ami.ec2.id
  instance_type = var.instance_type
  subnet_id = var.subnet_id
  user_data = var.user_data
  vpc_security_group_ids = [aws_security_group.ec2.id]
  associate_public_ip_address = var.eip || var.public_ip_address ? true : false
  key_name = var.key_name
  root_block_device {
    volume_size = var.volume_size
    volume_type = var.volume_type
    encrypted = true
  }

  tags = {
    "Name" = var.prefix
  }
  lifecycle {
     ignore_changes = [ user_data, ami ]
  }
}

resource "aws_route53_record" "ec2" {
  count = var.route53_zone_id != "" ? 1 : 0 
  zone_id = var.route53_zone_id
  name = var.route53_name == "thisisundefined" ? var.prefix : var.route53_name
  type = "A"
  ttl = 60
  records = var.eip ? [aws_eip.ec2[0].public_ip] : [aws_instance.ec2.public_ip]
}
