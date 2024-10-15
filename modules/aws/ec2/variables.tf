variable "prefix" {
  type = string
}

variable "ami_owners" {
  type = list(string)
  default = ["099720109477"]
}

variable "ami_architectures" {
  type = list(string)
  default = ["x86_64"]
}

variable "ami_names" {
  type = list(string)
  default = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
}

variable "instance_type" {
  type = string
  default = "t3.medium"
}

variable "subnet_id" {
  type = string
}

variable "key_name" {
  type = string
}

variable "volume_size" {
  type = number
  default = 200
}

variable "volume_type" {
  type = string
  default = "gp3"
}

variable "eip" {
  type = bool
  default = false
}

variable "public_ip_address" {
  type = bool
  default = false
}

variable "route53_zone_id" {
  type = string
  default = ""
}

variable "route53_name" {
  type = string
  nullable = false
  default = "thisisundefined"
}

variable "user_data" {
  type = string
  default = ""
}

variable "security_group_egress" {
  type = map(object({
    from_port = number
    to_port = number
    protocol = string
    cidr_blocks = list(string)
  }))
  default = {
    all = {
      from_port = 0
      to_port = 0
      protocol = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }
}

variable "security_group_ingress" {
  type = map(object({
    from_port = number
    to_port = number
    protocol = string
    cidr_blocks = list(string)
  }))
  default = {
    ssh = {
      from_port = 22
      to_port = 22
      protocol = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }
}
