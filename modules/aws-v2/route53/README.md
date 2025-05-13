## Standard domain structure with route53 ##
This module creates route53 zones (public and/or private) + adds ACM TLS certificates with validation for both or uses PCA.

vpc_id - set the default VPC to be used for private zones.

domains - configure the zones to be created. Map of objects.
```
object fields:
    domain_name - FQDN
    parent_zone_id - specify the parent zone ID where the NS records of the zone will be placed in. (default "", Optional)
    create_zone - create tehe zone when true, when false the zone will be fetched with data request. (default true, optional)
    create_certificate - will create teh ACM Certificate for the zone CN: "*.domain_name" and "domain_name". (default true, optional)
    certificate_key_algorithm  - algorithm to use for the ACM certificate (default "EC_secp384r1", optional)
    certificate_authority_arn - when using PCA then specify the ARN of the CA. (default "", optional)
    private - create the domain as a private zone (default false, optional)
    vpc_id - specify the vpc_id for the private zone. When not specified the var.vpc_id will be used. (default "", optional)
    default_public - set this as the default public domain to be used for other modules
    default_private - set this as the default private domain to be used for other modules
```

### Example code ###
Basic example:
```
        - name: dns
          source: aws-v2/route53
          inputs:
            domains: |
              {
                "public" = {
                  domain_name        = "example.entigo.com"
                },
                "private" = {
                  domain_name        = "example-int.entigo.com"
                  private            = true
                }
              }

```
Minimal example with existing domain:
```
domains: |
  {
    "min" = {
      domain_name        = "example.com"
      create_certificate = true
      create_zone        = false
    }
  }
```
Create the NS recods in parent zone and specify what are the defaults.
```
domains: |
  {
    "extpublic" = {
      domain_name        = "route53v2-ext.example.com"
      parent_zone_id     = "XXXXXXXXXXXXXXXXXXXXX"
      default_public     = true
    },
    "extprivate" = {
      domain_name        = "route53v2-ext-int.example.com"
      parent_zone_id     = "XXXXXXXXXXXXXXXXXXXXX"
      default_private    = true
    }
    "extpublic-extra" = {
      domain_name        = "route53v2-ext-extra.example.com"
      create_certificate = false
    }
  }
```
