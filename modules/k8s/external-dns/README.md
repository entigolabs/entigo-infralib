## Opinionated helm package for external-dns ##

In addition to installing external-dns with Helm it also created the needed IRSA with crossplane.



### Example code ###

```
    modules:
      - name: external-dns
        source: external-dns

```

### Azure ###

Azure private zones need their own provider, so on Azure a second instance `<module name>-external-dns-private`
(`azure-private-dns`, HTTPRoutes only) writes the private int zone. Both share one WorkloadIdentity (crossplane-azure):
DNS Zone Contributor on `pub_zone_id`, Private DNS Zone Contributor on `int_zone_id`. Set
`external-dns-private.enabled: false` when the int zone is not private.

### DNSEndpoint ###

The `crd` source is enabled, so records that no Service, Ingress or Route can
express can be created directly. The Gateway API source only takes the record
target from the Gateway, so a single hostname on a shared Gateway that has to
point somewhere other than the load balancer needs a DNSEndpoint:

```
apiVersion: externaldns.k8s.io/v1alpha1
kind: DNSEndpoint
metadata:
  name: example
spec:
  endpoints:
    - dnsName: example.mydomain.com
      recordType: CNAME
      recordTTL: 60
      targets:
        - example.mydomain.com.cdn.cloudflare.net
```

If another source also produces a record for the same hostname, the A record
wins and the CNAME is dropped, so keep the hostname off the Route (or annotate
the Route with `external-dns.kubernetes.io/controller: none`).
