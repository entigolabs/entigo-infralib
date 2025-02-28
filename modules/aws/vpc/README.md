## Oppinionated module for vpc creation ##


Oppinionated version of this https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest

### Automatic subnet calculations ###
If you do not specify the subnet ranges and only the vpc_cidr, then the module will create the subnets automatically. When the range is smaller than 19 (/21, /22 ...) then the database and elasticache subnets are not created.

We do leave a lot of spare space when using (azs = 2). This is done to enable clients to later switch to a 3 zone setup without recreating subnets. It will also keep NACL rules simple - by allowing one range to cover the same type subnets.

If the automatic calculation is not to Your liking then specify each subnet as desired.

#### vpc_cidr: "10.24.0.0/16"
![10.24.0.0/16](test/size-16.png)

#### vpc_cidr: "10.24.0.0/18"
![10.24.0.0/18](test/size-18.png)

#### vpc_cidr: "10.24.0.0/19"
![10.24.0.0/19](test/size-19.png)

#### vpc_cidr: "10.24.0.0/20"
![10.24.0.0/20](test/size-20.png)

#### vpc_cidr: "10.24.0.0/21"
![10.24.0.0/21](test/size-21.png)

#### vpc_cidr: "10.24.0.0/24"
![10.24.0.0/24](test/size-24.png)


### Example code ###
```
    modules:
      - name: vpc
        source: aws/vpc
        inputs:
          vpc_cidr: "10.24.0.0/20"
          one_nat_gateway_per_az = false #Will only create 1 nat gw, not two
          azs = 2
```
Will result in networks:
__public ( 10.24.0.0/22 for NACL )__

public-a ( 10.24.0.0/24 )

public-b ( 10.24.1.0/24 )

__intra ( 10.24.4.0/22 for NACL )__

intra-a ( 10.24.4.0/24 )

intra-b ( 10.24.5.0/24 )


__private ( 10.24.8.0/21 for NACL )__

private-a ( 10.24.8.0/23 )

private-b ( 10.24.10.0/23 )

__No database networks are created.__

__No elasticsearch networks are created.__



```
    modules:
      - name: vpc
        source: aws/vpc
        inputs:
          vpc_cidr: "10.24.0.0/20"
          one_nat_gateway_per_az = true #Will create three nat gw-s into each public subnet
          azs = 3
          intra_subnets: |
            ["10.24.4.0/23", "10.24.6.0/23"]
          database_subnets: |
            ["10.24.14.0/24", "10.24.15.0/24"]

```
Will result in networks:
__public ( 10.24.0.0/22 for NACL )__

public-a ( 10.24.0.0/24 )

public-b ( 10.24.1.0/24 )

public-c ( 10.24.3.0/24 )

__intra ( 10.24.4.0/22 for NACL )__

intra-a ( 10.24.4.0/23 )

intra-b ( 10.24.6.0/23 )


__private ( 10.24.8.0/21 for NACL )__

private-a ( 10.24.8.0/23 )

private-b ( 10.24.10.0/23 )

private-c ( 10.24.10.0/23 )

__database ( 10.24.14.0/23 for NACL )__

db-a 10.24.14.0/24

db-b 10.24.15.0/24

__No elasticsearch networks are created.__
