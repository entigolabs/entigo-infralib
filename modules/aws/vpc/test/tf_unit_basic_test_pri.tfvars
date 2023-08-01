vpc_cidr = "10.146.0.0/16"

one_nat_gateway_per_az = false
private_subnets = ["10.146.32.0/21", "10.146.40.0/21", "10.146.48.0/21"]
public_subnets = ["10.146.4.0/24", "10.146.5.0/24"]
intra_subnets = ["10.146.0.0/26"]
