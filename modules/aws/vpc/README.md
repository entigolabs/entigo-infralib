## Oppinionated module for vpc creation ##


Oppinionated version of this https://registry.terraform.io/modules/terraform-aws-modules/vpc/aws/latest


### SSM parameters ###
```
"/entigo-infralib/${local.hname}/vpc/pipeline_security_group"
"/entigo-infralib/${local.hname}/vpc/vpc_id"
"/entigo-infralib/${local.hname}/vpc/private_subnets"
"/entigo-infralib/${local.hname}/vpc/public_subnets"
"/entigo-infralib/${local.hname}/vpc/intra_subnets"
"/entigo-infralib/${local.hname}/vpc/database_subnets"
"/entigo-infralib/${local.hname}/vpc/database_subnet_group"
"/entigo-infralib/${local.hname}/vpc/elasticache_subnets"
"/entigo-infralib/${local.hname}/vpc/elasticache_subnet_group"
"/entigo-infralib/${local.hname}/vpc/private_subnet_cidrs"
"/entigo-infralib/${local.hname}/vpc/public_subnet_cidrs"
"/entigo-infralib/${local.hname}/vpc/database_subnet_cidrs"
"/entigo-infralib/${local.hname}/vpc/elasticache_subnet_cidrs"
"/entigo-infralib/${local.hname}/vpc/intra_subnet_cidrs"

```


### Example code ###

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
