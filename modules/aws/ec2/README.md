## Create EC2 instance ##



__ami_owners__  default = ["099720109477"]

__ami_architectures__ default = ["x86_64"]

__ami_names__ default = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]

__instance_type__ default = "t3.medium"

__subnet_id__ the subnet ID to where the ec2 instance will be created

__key_name__ ssh key name

__volume_size__ default = 200 root volume size in GB

__volume_type__ default = "gp3"

__eip__ default = false create elastic IP boolean

__public_ip_address__ default = false assign public IP boolean

__route53_zone_id__ Create DNS record to this zone (at the moment only supports instance with public IP)

__user_data__ run commands on instance creation

__security_group_egress__ at the moment only support CIDR blocks
```
  default = {
    all = {
      from_port = 0
      to_port = 0
      protocol = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }
```

__security_group_ingress__ at the moment only support CIDR blocks
```
  default = {
    ssh = {
      from_port = 22
      to_port = 22
      protocol = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }
```

### Example code ###

```
...

```
