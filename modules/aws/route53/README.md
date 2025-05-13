## Deprecated ##
This module is no longer actively developed, please use aws-v2/route53 module instead.

## Migration from aws/route53 to aws-v2/route53-v2

In this migration the route53 module has a name "dns" and it is in the "infra" step. The terraform commands will have to be changed to suite your configuration.

Example original configuration:
```
        - name: dns
          source: aws-v2/route53
          inputs:
            parent_domain: entigo.dev
            private_subdomain_name: dev-int
            public_subdomain_name: dev
            vpc_id: '{{ ssm.net.main.vpc_id }}'
```
Example new configuration:
```
        - name: dns
          source: aws-v2/route53
          inputs:
            domains: |
              {
                "public" = {
                  domain_name        = "dev.entigo.dev"
                  certificate_key_algorithm      = "RSA_2048"
                },
                "private" = {
                  domain_name        = "dev-int.entigo.dev"
                  private            = true
                  certificate_key_algorithm      = "RSA_2048"
                }
              }
```

Run the agent and DISCARD/Reject the plan for the changes. Make a backup of the state file in S3.
```
docker run --pull always -it --rm -v "$(pwd)":"/etc/ei-agent" -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e AWS_REGION=$AWS_REGION -e AWS_SESSION_TOKEN=$AWS_SESSION_TOKEN -e CONFIG=/etc/ei-agent/config.yaml entigolabs/entigo-infralib-agent ei-agent run --steps infra 
```

Move the aws_acm_certificate, aws_route53_zone and aws_route53_record resources to new locations.
```
docker run -it --rm -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e AWS_REGION=$AWS_REGION -e AWS_SESSION_TOKEN=$AWS_SESSION_TOKEN --entrypoint /bin/bash entigolabs/entigo-infralib-base
aws s3 cp s3://<CONFIG PREFIX>-<ACCOUNT NUMBER>-$AWS_REGION/steps/<CONFIG PREFIX>-<STEP NAME> ./tmp --recursive --exclude "*.terraform/*"
cd ./tmp
terraform init -input=false -backend-config=backend.conf
terraform plan #The plan will show that it want to destroy many resources. 
terraform state list | grep aws_acm_certificate
terraform state mv 'module.dns.aws_acm_certificate.int[0]' 'module.dns.aws_acm_certificate.this["private"]'
terraform state mv 'module.dns.aws_acm_certificate.pub[0]' 'module.dns.aws_acm_certificate.this["public"]'
terraform state list | grep aws_acm_certificate

terraform state list | grep aws_route53_zone
terraform state mv 'module.dns.aws_route53_zone.int[0]' 'module.dns.aws_route53_zone.this["private"]'
terraform state mv 'module.dns.aws_route53_zone.int-cert[0]' 'module.dns.aws_route53_zone.validation["private"]'
terraform state mv 'module.dns.aws_route53_zone.pub[0]' 'module.dns.aws_route53_zone.this["public"]'
terraform state list | grep aws_route53_zone

terraform state list | grep aws_route53_record
terraform state mv 'module.dns.aws_route53_record.int-cert["*.dev-int.entigo.dev"]' 'module.dns.aws_route53_record.validation["private_*.dev-int.entigo.dev"]'
terraform state mv 'module.dns.aws_route53_record.int-cert["dev-int.entigo.dev"]' 'module.dns.aws_route53_record.validation["private_dev-int.entigo.dev"]'
terraform state mv 'module.dns.aws_route53_record.pub-cert["*.dev.entigo.dev"]' 'module.dns.aws_route53_record.validation["public_*.dev.entigo.dev"]'
terraform state mv 'module.dns.aws_route53_record.pub-cert["dev.entigo.dev"]' 'module.dns.aws_route53_record.validation["public_dev.entigo.dev"]'
terraform state list | grep aws_route53_record

terraform plan #Now the plan should only show changes to tags and nothing to destroy. If not find what resources are still mismatchin or what config changes cause a destructive plan.
exit
```
Now run the agent again and verify the plan is not going to destroy your zones or certificates.


## Standard domain structure with route53 ##
This module creates route53 zones (public and/or private) + adds ACM TLS certificates with validation for both.

You need to specify either __parent_zone_id__(existing AWS Zone ID in the same account) or __parent_domain__(DNS Nameservers are configured manually).
If you set __parent_domain__ then the certificates are created but not validated.

__create_public__ defalts to true, but if set to false then the parent zone is used insted and a public route53 zone is not created. The zone name will be ${prefix}-${stepname}-${modulename}.${parent-domain-name} by default.

__public_subdomain_name__ enables to override the default name of the public subdomain. The parent domain is automatically added.

__create_private__ default to true, but if set to false then the public or parent zone is used and a private route53 zone is not created. The zone name will be ${prefix}-${stepname}-${modulename}-int.${parent-domain-name} by default.

__private_subdomain_name__ enables to override the default name of the private subdomain. The parent domain is automatically added.

__create_cert__ defaults to true, but if set to false then no ACM certificates are created. The private domain will also get an equal public domain so the ACM Certificate could be validated.

__vpc_id__ need to reference the VPC that the private dns zone will be attatched to. Only requied if __create_private__=true.

The use of the private zone only makes sense if we have LAN access to that network (for example Client VPN or access through TGW).


### Example code ###

```
    modules:
      - name: dns
        inputs:
          parent_zone_id: "Z0798XXXXXXXXXXXXXXXX"

```
Or 
```
    modules:
      - name: dns
        inputs:
          create_private: false
          parent_domain: "entigo.io"

```
