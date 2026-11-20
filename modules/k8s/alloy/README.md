# alloy

Grafana Alloy as the collector of the platform. It ships logs to Loki and, when
enabled, metrics to Mimir. Three installations of the upstream chart:

| Part | Kind | Default | Does |
|---|---|---|---|
| `node-agent` | DaemonSet | on | Pod logs to Loki. With `metrics.enabled`: node metrics (the built-in node exporter), kubelet and cAdvisor of its own node. A node-local OTLP receiver for traces when `traces.enabled`. |
| `cluster-metrics` | Deployment | off | Everything that is scraped once per cluster: kube-state-metrics, pods and services with the `prometheus.io/scrape` annotation, the API server, optionally ServiceMonitors, PodMonitors and PrometheusRules. |
| `receiver` | Deployment | off | Push receivers (Faro, OTLP) for producers outside the cluster. |

With the default inputs the module is a log shipper only.

```yaml
    modules:
      - name: alloy
        source: alloy
```

## Metrics

Use either this module's metrics or the `prometheus` module on a cluster, not
both. They scrape the same targets, so running both collects everything twice,
and both can install the Prometheus Operator CRDs.

Turn the two parts on together. `node-agent` covers what lives on a node,
`cluster-metrics` covers the rest, and the bundled Grafana dashboards need both:

```yaml
      - name: alloy
        source: alloy
        inputs:
          node-agent:
            metrics:
              enabled: true
          cluster-metrics:
            enabled: true
```

What this needs next to it:

- **A Mimir to write to.** By default the `mimir` module of the same platform,
  found by the agent. For a Mimir somewhere else see
  [Mimir in another place](#mimir-in-another-place).
- **The `kube-state-metrics` module.** The `prometheus` module brings its own
  kube-state-metrics, this module does not. `cluster-metrics` scrapes the pods
  labelled `app.kubernetes.io/name: kube-state-metrics` on port 8080; without
  the module the `kube_*` metrics, and every dashboard built on them, are empty.

```yaml
      - name: kube-state-metrics
        source: kube-state-metrics
      - name: mimir
        source: mimir
```

Every series gets a `cluster` label with the name of the cluster, set by the
agent. All clusters write to the same Mimir tenant, `anonymous`, and are told
apart by that label. Change the tenant with `node-agent.metrics.mimirTenantID`
and `cluster-metrics.mimirTenantID`, and the labels with
`node-agent.metrics.externalLabels` and `cluster-metrics.externalLabels`.

### Rules, ServiceMonitors and the CRDs

`cluster-metrics` can work with the Prometheus Operator resources that many
charts ship. Each is off by default:

```yaml
        inputs:
          global:
            installPrometheusOperatorCRDs: true
          cluster-metrics:
            enabled: true
            rules: true
            serviceMonitors: true
            podMonitors: true
```

| Input | Does |
|---|---|
| `global.installPrometheusOperatorCRDs` | Installs the Prometheus Operator CRDs (ServiceMonitor, PodMonitor, PrometheusRule, ...). Needed by the three below unless something else already installs them. Only the CRDs, no operator. |
| `cluster-metrics.rules` | Copies every `PrometheusRule` of the cluster into the Mimir ruler, which evaluates the recording and alerting rules. |
| `cluster-metrics.serviceMonitors` | Scrapes what the `ServiceMonitor` resources of the cluster select. |
| `cluster-metrics.podMonitors` | Scrapes what the `PodMonitor` resources of the cluster select. |

Alerting rules fire into Mimir's Alertmanager, which the `mimir` module leaves
off by default. Without it the rules are evaluated but no alert is delivered:

```yaml
      - name: mimir
        source: mimir
        inputs:
          mimir-distributed:
            alertmanager:
              enabled: true
```

### Mimir in another place

When Mimir runs in another cluster or another platform, set the remote write
URL of both parts. An input replaces what the agent would have looked up:

```yaml
        inputs:
          node-agent:
            metrics:
              enabled: true
              mimirEndpoint: "http://mimir.example.internal/api/v1/push"
          cluster-metrics:
            enabled: true
            mimirEndpoint: "http://mimir.example.internal/api/v1/push"
```

- The URL ends in `/api/v1/push`. With `cluster-metrics.rules` the rules go to
  the same host.
- Set `mimirTenantID` on both parts when that Mimir expects another tenant
  than `anonymous`.
- The NetworkPolicies of both parts allow egress to ports 80 and 8080 on any
  address. An endpoint on another port, 443 for example, needs that port
  added to `node-agent.networkPolicy.egress` and
  `cluster-metrics.networkPolicy.egress`. These are lists, so the input has to
  repeat the whole list of `agent_input_<cloud>.yaml` with the rule added.

Loki works the same way, with `node-agent.logs.lokiEndpoint` (ending in
`/loki/api/v1/push`) and `node-agent.logs.lokiTenantID`.
