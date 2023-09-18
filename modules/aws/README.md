## Terraform modules that are specific to AWS ##

__nuke.sh__  if runs locally then will first show what will be destroyed in entigo-infralib AWS account and then promts for confirmation. if runs in github actions then it will not promt and destroys all resources.
This helps

__aws-nuke-config.yml__ configuration of AWS Nuke - mostly needed to exclude some resources that won't be nuked every day in entigo-infralib AWS account.


These modules can be used in the entigo-agent steps of "__type: terraform__"

## Example code ##
```
steps:
  - name: network
    type: terraform
    workspace: test
    approve: minor
    modules:
      - name: vpc
        source: aws/vpc
        version: stable
        inputs:
          vpc_cidr: "10.175.0.0/16"
          one_nat_gateway_per_az: true
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
