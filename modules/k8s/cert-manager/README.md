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

* AWS Private CA: `global.aws.privateCA.enabled: true` adds the aws-privateca-issuer subchart, so certificates can be issued
  from AWS Private CA, also from a CA of another account shared with AWS RAM. `global.aws.privateCA.arns` lists the CAs: the
  IRSA role of the issuer (crossplane-aws) may issue from those only, and the `AWSPCAClusterIssuer` `aws-privateca`
  is created for the first one. `global.aws.privateCA.clusterIssuer.create: false` installs the issuer controller alone, for
  own `AWSPCAIssuer` and `AWSPCAClusterIssuer` objects. Depends on: modules/k8s/crossplane-aws.

Another cloud adds `values-<cloud>.yaml` (`global.cloudProvider`), its DNS module's zone outputs to the
`global.pubZoneId` and `global.intCertZoneId` chains in `agent_input.yaml`, and `templates/<cloud>/` with its
ClusterIssuer (ACME account plus its DNS-01 solvers) and the DNS identity of the controller.

Test environments should use `acme.server: https://acme-staging-v02.api.letsencrypt.org/directory`: production
allows 5 certificates per identical name set per 7 days.

### Issuers ###

* `clusterIssuer.create` creates the ACME ClusterIssuer and the cloud identity of the controller; `clusterIssuer.name`
  and `acme.server` select the ACME CA. CAs that need External Account Binding (ZeroSSL, Sectigo, Google Public CA):
  `acme.externalAccountBinding` with a Secret holding the HMAC key in this namespace.
* `global.aws.privateCA` (AWS): the `AWSPCAClusterIssuer` `global.aws.privateCA.clusterIssuer.name` (default `aws-privateca`). A
  Certificate refers to it with `issuerRef` group `awspca.cert-manager.io`, kind `AWSPCAClusterIssuer`. The usages of
  the Certificate select the AWS Private CA template, and a CA shared with the default AWS RAM permission only allows
  the default end entity template (`server auth` together with `client auth`).
* `extraClusterIssuers`: more ClusterIssuers with the spec as is (CA, Vault, ...).
* azure-gateway and istio-gateway default to `clusterIssuer.name` of this module, their `certificateIssuer` picks
  another one.

### Updating aws-privateca-issuer ###

The subchart keeps its CRDs in `crds/`, which ArgoCD does not render (modules/k8s/README.md), so they are copied into
`templates/aws/privateCACrds.yaml`. After a version bump of the `aws-privateca-issuer` dependency copy the two files of
the new tarball's `crds/` there again, inside the `global.aws.privateCA.enabled` gate, with the annotations
`helm.sh/resource-policy: keep` and `argocd.argoproj.io/sync-wave: "-1"` on each.

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

AWS with an AWS Private CA issuer:

```
    modules:
      - name: cert-manager
        source: cert-manager
        inputs:
          global:
            aws:
              privateCA:
                enabled: true
                arns:
                  - arn:aws:acm-pca:eu-north-1:111122223333:certificate-authority/11111111-2222-3333-4444-555555555555
```

A Certificate issued from the AWS Private CA issuer:

```
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-pca
spec:
  secretName: example-pca-certificate
  commonName: example-pca.your.domain
  dnsNames:
    - example-pca.your.domain
    - foo-example-pca.your.domain
  duration: 2160h      # 90 days
  renewBefore: 720h    # renew 30 days before expiry
  privateKey:
    algorithm: ECDSA
    size: 384
    rotationPolicy: Always
  usages:
    - server auth
    - client auth
  issuerRef:
    group: awspca.cert-manager.io
    kind: AWSPCAClusterIssuer
    name: aws-privateca
```
