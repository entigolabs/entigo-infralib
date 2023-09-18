## Standard domain structure with route53 ##
This module creates route53 zones (public and/or private) + adds ACM TLS certificates with validation for both.

You need to specify either __parent_zone_id__(existing AWS Zone ID in the same account) or __parent_domain__(DNS Nameservers are configured manually).

__create_public__ defalts to true, but if set to false then the private zone is used insted of public and a public route53 zone is not created. The zone name will be ${prefix}-${stepname}-${modulename}.${parent-domain-name}

__create_private__ default to true, but if set to false then the existing parent zone is used and a private  route53 zone is not created. The zone name will be ${prefix}-${stepname}-${modulename}-int.${parent-domain-name}

__create_cert__ defaults to true, but if set to false then no ACM certificates are created. The private domain will also get an equal public domain so the ACM Certificate could be validated.

vpc_prefix need to reference the VPC that the private dns zone will be attatched to. (prefix + step name + module name) Only needed if create_private=true.

The use of the private zone only makes sense if we have LAN access to that network (for example Client VPN or access through TGW).


### SSM parameters ###
```
"/entigo-infralib/${local.hname}/route53/pub_zone_id" 
"/entigo-infralib/${local.hname}/route53/pub_domain"
"/entigo-infralib/${local.hname}/route53/int_zone_id"
"/entigo-infralib/${local.hname}/route53/int_domain"
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
