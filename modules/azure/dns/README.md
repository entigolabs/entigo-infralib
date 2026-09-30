## Azure DNS zones ##

Creates Azure DNS zones like aws-v2/route53: public zones and private zones (linked to `vpc_ids`, default the vpc
module's VNet). Private zones get a public twin with the same name for ACME DNS-01 challenges (`create_validation`);
certificates are issued by the cert-manager k8s module, not here.

domains - map of objects:
```
    domain_name       - FQDN
    parent_zone_id    - Azure DNS zone resource id, NS records are created there when it is in the same resource
                        group, otherwise add them manually from the nameservers / validation_nameservers outputs
                        (default "")
    create_zone       - false uses an existing zone in the resource group (default true)
    create_validation - public validation zone for a private domain (default true)
    private           - private DNS zone (default false)
    vpc_ids           - VNets linked to the private zone, defaults to var.vpc_ids
    default_public    - the default public domain for other modules (pub_zone_id, pub_domain)
    default_private   - the default private domain for other modules (int_zone_id, int_domain)
```
With a single domain, or a single domain of a kind, the defaults are set automatically. Exactly one domain must be
the default public and one the default private one; a single public domain is both.

### Example code ###
```
    modules:
      - name: dns
        source: azure/dns
        inputs:
          domains: |
            {
              "public" = {
                domain_name = "example.entigo.com"
              },
              "private" = {
                domain_name = "example-int.entigo.com"
                private     = true
              }
            }
```
