## cert-manager with Let's Encrypt ##

Installs cert-manager and (when `issuer.ingressClassName` is set) a `letsencrypt`
ClusterIssuer using HTTP-01 - the challenge is served through the cloud's ingress
controller, so no DNS-provider credentials or webhook solvers are needed. Used on
Oracle to give `modules/k8s/argocd` a publicly trusted certificate, terminated on
the OCI LB by oci-native-ingress-controller.

### Example code ###

```
    modules:
      - name: cert-manager
        source: cert-manager
```
