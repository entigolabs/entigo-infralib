## Opinionated helm package for cert-manager ##

cert-manager with an ACME `ClusterIssuer` (DNS-01 in the cloud DNS), default `letsencrypt`. One per cluster.

* Azure (AzurePublicCloud): Application Gateway for Containers reads certificates only from Kubernetes Secrets, so the
  issuer is created. The controller identity is a WorkloadIdentity (crossplane-azure) with DNS Zone Contributor on the
  dns `pub_zone_id` and `int_cert_zone_id` zones. Depends on: modules/azure/dns, modules/k8s/crossplane-azure.
* AWS: the gateways use ACM certificates, so by default cert-manager is installed alone, for charts that bring their
  own issuers (kcp, operator webhooks). `clusterIssuer.create: true` adds the issuer with Route53 DNS-01 solvers for
  the `pub_zone_id` and `int_cert_zone_id` (public validation twin of the private int zone) zones of a route53 module
  named `route53`, and an IRSA role (crossplane-aws) that may change records in those two zones only. Depends on:
  modules/aws-v2/route53, modules/k8s/crossplane-aws.

Another cloud adds `values-<cloud>.yaml` (`global.cloudProvider`), its DNS module's zone outputs to the
`global.pubZoneId` and `global.intCertZoneId` chains in `agent_input.yaml`, and `templates/<cloud>/` with its
ClusterIssuer (ACME account plus its DNS-01 solvers) and the DNS identity of the controller.

Test environments should use `acme.server: https://acme-staging-v02.api.letsencrypt.org/directory`: production
allows 5 certificates per identical name set per 7 days.

### Issuers ###

* `clusterIssuer.create` creates the ACME ClusterIssuer and the cloud identity of the controller; `clusterIssuer.name`
  and `acme.server` select the ACME CA. CAs that need External Account Binding (ZeroSSL, Sectigo, Google Public CA):
  `acme.externalAccountBinding` with a Secret holding the HMAC key in this namespace.
* `extraClusterIssuers`: more ClusterIssuers with the spec as is (CA, Vault, ...).
* azure-gateway and istio-gateway default to `clusterIssuer.name` of this module, their `certificateIssuer` picks
  another one.

### Example code ###

```
    modules:
      - name: cert-manager
        source: cert-manager
```

AWS with the ACME issuer:

```
    modules:
      - name: cert-manager
        source: cert-manager
        inputs:
          clusterIssuer:
            create: true
          acme:
            email: ops@example.com
```
