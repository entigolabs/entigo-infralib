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


### SSM parameters ###
```
"/entigo-infralib/${local.hname}/pub_zone_id" 
"/entigo-infralib/${local.hname}/pub_domain"
"/entigo-infralib/${local.hname}/int_zone_id"
"/entigo-infralib/${local.hname}/int_domain"
```


### Example code ###

```
    modules:
      - name: dns
        inputs:
          vpc_prefix: "ep-network-vpc"
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
