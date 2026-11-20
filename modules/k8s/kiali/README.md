# kiali

Kiali, the console of the Istio service mesh. It draws its graphs from the
Istio metrics in a Prometheus-compatible store.

```yaml
    modules:
      - name: kiali
        source: kiali
```

## Where the metrics come from

By default Kiali queries the `prometheus` module of the same platform, which
the agent looks up. That lookup is required: on a platform without a
`prometheus` module the agent stops with
`failed to find module with type prometheus`, unless the URL is set in the
inputs. An input replaces the lookup.

### Mimir instead of Prometheus

For a platform that keeps its metrics in Mimir (the `alloy` module with
metrics enabled, see [alloy](../alloy/README.md#metrics)):

```yaml
      - name: kiali
        source: kiali
        inputs:
          kiali-server:
            external_services:
              prometheus:
                url: "http://mimir-gateway.{{ .tmodule.mimir }}/prometheus"
                # Kiali asks Prometheus for its scrape interval and retention. Mimir does
                # not answer that, so they are stated here.
                thanos_proxy:
                  enabled: true
                  scrape_interval: "60s"
                  retention_period: "7d"
```

- The URL is the Mimir gateway with the `/prometheus` path.
- No tenant header is needed for the default tenant: the Mimir gateway puts a
  request without `X-Scope-OrgID` into `anonymous`, the tenant Alloy writes to.
  For another tenant add
  `custom_headers: { X-Scope-OrgID: "<tenant>" }` next to `url`.
- `scrape_interval` is Alloy's (`60s` unless changed), `retention_period` how
  long that Mimir keeps metrics.
- When several clusters write to the same Mimir tenant, limit Kiali to its own
  cluster, otherwise their traffic is mixed into one graph:

```yaml
                query_scope:
                  cluster: "<the cluster label of this cluster>"
```

### Another Prometheus

A Prometheus that is not the `prometheus` module of this platform, in the
cluster or outside it:

```yaml
        inputs:
          kiali-server:
            external_services:
              prometheus:
                url: "http://prometheus.example.internal:9090"
```

Kiali has to be able to reach it, and it has to hold the Istio metrics of this
cluster (`istio_requests_total` and the rest of the sidecar and istiod metrics).
