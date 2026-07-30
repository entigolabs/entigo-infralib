## ingress-nginx ##

Community ingress-nginx controller behind a cloud TCP load balancer. TLS terminates
in nginx (full SNI), so every app brings its own cert-manager certificate - on Oracle
this sidesteps the OCI Load Balancer's one-certificate-per-listener limit that makes
the OCI Native Ingress Controller (modules/k8s/oci-native-ingress-controller)
unsuitable for multiple HTTPS apps behind one LB.

### Example code ###

```
    modules:
      - name: ingress-nginx
        source: ingress-nginx
```
