## Oppinionated module for vpc creation ##


Oppinionated version of this https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest






### SSM parameters ###
```
"/entigo-infralib/${local.hname}/pipeline_security_group"
"/entigo-infralib/${local.hname}/vpc_id"
"/entigo-infralib/${local.hname}/private_subnets"
"/entigo-infralib/${local.hname}/public_subnets"
"/entigo-infralib/${local.hname}/intra_subnets"
"/entigo-infralib/${local.hname}/database_subnets"
"/entigo-infralib/${local.hname}/database_subnet_group"
"/entigo-infralib/${local.hname}/elasticache_subnets"
"/entigo-infralib/${local.hname}/elasticache_subnet_group"
"/entigo-infralib/${local.hname}/private_subnet_cidrs"
"/entigo-infralib/${local.hname}/public_subnet_cidrs"
"/entigo-infralib/${local.hname}/database_subnet_cidrs"
"/entigo-infralib/${local.hname}/elasticache_subnet_cidrs"
"/entigo-infralib/${local.hname}/intra_subnet_cidrs"

```


### Example code ###
```
    modules:
      - name: vpc
        source: aws/vpc
        inputs:
          vpc_cidr: "10.24.16.0/21"
          one_nat_gateway_per_az = false #Will only create 1 nat gw, not two
          elasticache_subnets = []
          azs = 2
```
Will result in networks:
__public ( 10.24.16.0/24 for NACL )__

public-a ( 10.24.16.0/26 )

public-b ( 10.24.16.64/26 )

public-spare ( 10.24.16.128/26 )

public-spare ( 10.24.16.192/26 )

__intra ( 10.24.17.0/24 for NACL )__

intra-a ( 10.24.17.0/26 )

intra-b ( 10.24.17.64/26 )

intra-spare ( 10.24.17.128/26 )

intra-spare ( 10.24.17.192/26 )

__private ( 10.24.18.0/23 for NACL )__

intra-a ( 10.24.18.0/25 )

intra-b ( 10.24.18.128/25 )

intra-spare ( 10.24.19.0/25 )

intra-spare ( 10.24.19.128/25 )

__database ( 10.24.20.0/23 for NACL )__

database-a ( 10.24.20.0/25 )

database-b ( 10.24.20.128/25 )

database-spare ( 10.24.21.0/25 )

database-spare ( 10.24.21.128/25 )

__No elasticsearch networks are created.__

__spares:__ 

10.24.22.0/23

10.24.24.0/23

```
    modules:
      - name: vpc
        source: aws/vpc
        inputs:
          vpc_cidr: "10.24.16.0/21"
          one_nat_gateway_per_az = true #Will create three nat gw-s into each public subnet
          intra_subnets = ["10.24.17.0/24"]
          azs = 3
```
Will result in networks:
__public ( 10.24.16.0/24 for NACL )__

public-a ( 10.24.16.0/26 )

public-b ( 10.24.16.64/26 )

public-c ( 10.24.16.128/26 )

public-spare ( 10.24.16.192/26 )


__intra ( 10.24.17.0/24 for NACL ) (only in 1 AZ)__

intra-a ( 10.24.17.0/24 ) (only in 1 AZ)


__private ( 10.24.18.0/23 for NACL )__

intra-a ( 10.24.18.0/25 )

intra-b ( 10.24.18.128/25 )

intra-c ( 10.24.19.0/25 )

intra-spare ( 10.24.19.128/25 )

__database ( 10.24.20.0/23 for NACL )__

database-a ( 10.24.20.0/25 )

database-b ( 10.24.20.128/25 )

database-c ( 10.24.21.0/25 )

database-spare ( 10.24.21.128/25 )

__elasticache ( 10.24.22.0/24 for NACL )__

elasticache-a ( 10.24.22.0/26 )

elasticache-b ( 10.24.22.64/26 )

elasticache-c ( 10.24.22.128/26 )

elasticache-spare ( 10.24.22.192/26 )

__spares:__ 

10.24.23.0/24

10.24.24.0/23


```
    modules:
      - name: vpc
        source: aws/vpc
        inputs:
          vpc_cidr: "10.175.0.0/16"
          one_nat_gateway_per_az: true #If set to true then all zones will have a nat gateway,otherwise only one public
          private_subnets: |
            ["10.175.32.0/21", "10.175.40.0/21", "10.175.48.0/21"]
          public_subnets: |
            ["10.175.4.0/24", "10.175.5.0/24", "10.175.6.0/24"]
          database_subnets: |
            ["10.175.16.0/22", "10.175.20.0/22", "10.175.24.0/22"]
          elasticache_subnets: |
            ["10.175.0.0/26", "10.175.0.64/26", "10.175.0.128/26"]
          intra_subnets: |
            []

```
Will create the network as described. If the entire network is custom then all the types should be specified (private_subnets, public_subnets, database_subnets, elasticache_subnets,intra_subnets) at least as empty lists "[]".
