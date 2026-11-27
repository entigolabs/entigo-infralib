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

Collect metrics with either this module or the `prometheus` module's own
scrape jobs on a cluster, not both. They scrape the same targets, so running
both collects everything twice, and both can install the Prometheus Operator
CRDs.

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

- **A store to write to.** By default the `mimir` module of the same platform,
  found by the agent, and the `prometheus` module when there is one, see
  [Prometheus as the store](#prometheus-as-the-store). For a Mimir somewhere
  else see [Mimir in another place](#mimir-in-another-place).
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

### What is left out

Two sources produce far more series than anything uses, so part of them is
dropped before it is sent. Each list is a regex on the metric name and can be
changed or emptied:

| Input | Default | Why |
|---|---|---|
| `node-agent.metrics.cadvisorDropMetrics` | a short list of per-container cAdvisor metrics (`container_tasks_state`, `container_memory_failures_total`, file system internals, ...) | No bundled dashboard uses them. |
| `cluster-metrics.apiserverDropMetrics` | `.+_bucket` | The API server's histogram buckets are the largest source of series in a cluster. Their `_sum` and `_count` stay, so rates and averages work; percentiles do not. |

To keep the API server's request latency percentiles, list the histograms to
drop instead of all of them, for example
`apiserver_(response_sizes|watch_events_sizes|admission_.*|flowcontrol_.*)_bucket|etcd_request_duration_seconds_bucket|workqueue_.*_bucket`.

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
- The NetworkPolicies allow egress to ports 80 and 8080 on any address, and
  `cluster-metrics` also to 443. An endpoint on another port, or on 443 for
  `node-agent`, needs that port added to `node-agent.networkPolicy.egress` or
  `cluster-metrics.networkPolicy.egress`. These are lists, so the input has to
  repeat the whole list (`agent_input_<cloud>.yaml` for `node-agent`,
  `values.yaml` for `cluster-metrics`) with the rule added.

### Endpoints behind OAuth2

A Loki or Mimir that accepts only OAuth2-authenticated pushes, such as the
Entigo portal's, takes client credentials next to its URL. Alloy fetches a token
from the token URL and sends it with every push:

```yaml
        inputs:
          node-agent:
            logs:
              lokiEndpoint: "https://portal.example.com/loki/api/v1/push"
              lokiOAuth2:
                clientId: "agent-example"
                clientSecret: "{{ .output-custom.agent-client-secret }}"
                tokenUrl: "https://idp.example.com/oauth/v2/token"
                scopes:
                - openid
            metrics:
              enabled: true
              mimirEndpoint: "https://portal.example.com/api/v1/push"
              mimirOAuth2:
                clientId: "agent-example"
                clientSecret: "{{ .output-custom.agent-client-secret }}"
                tokenUrl: "https://idp.example.com/oauth/v2/token"
                scopes:
                - openid
          cluster-metrics:
            enabled: true
            mimirEndpoint: "https://portal.example.com/api/v1/push"
            mimirOAuth2:
              clientId: "agent-example"
              clientSecret: "{{ .output-custom.agent-client-secret }}"
              tokenUrl: "https://idp.example.com/oauth/v2/token"
              scopes:
              - openid
```

- The credentials apply only when `clientId` is set.
- `cluster-metrics.mimirOAuth2` also authenticates the rules of
  `cluster-metrics.rules`, and `receiver.faro.lokiOAuth2` the logs that Faro
  takes in.
- With either OAuth2 input of `node-agent` set, the module adds a NetworkPolicy
  that lets the node agent reach port 443 on any address, for the token URL and
  the endpoint. `cluster-metrics` and `receiver` reach 443 already.

### Prometheus as the store

With metrics enabled, alloy writes to the Mimir and to the Prometheus of the
platform, whichever of the two modules exist, and to both when both do. The
agent finds them; nothing has to be set.

A Prometheus that alloy writes to is a store, not a collector. It needs its
remote write receiver on (`--web.enable-remote-write-receiver`), and its own
scrape jobs, node exporter and kube-state-metrics off, or everything is stored
twice. Enabling alloy metrics on a platform is the decision to stop Prometheus
scraping there.

The `prometheus` module has no single switch for this, the settings belong to
its subcharts. These inputs turn it into a store that still scrapes itself:

```yaml
      - name: prometheus
        source: prometheus
        inputs:
          prometheus:
            server:
              extraFlags:
                - web.enable-lifecycle
                - web.enable-remote-write-receiver
            scrapeConfigs:
              kubernetes-api-servers:
                enabled: false
              kubernetes-nodes:
                enabled: false
              kubernetes-nodes-cadvisor:
                enabled: false
              kubernetes-service-endpoints:
                enabled: false
              kubernetes-service-endpoints-slow:
                enabled: false
              prometheus-pushgateway:
                enabled: false
              kubernetes-services:
                enabled: false
              kubernetes-pods:
                enabled: false
              kubernetes-pods-slow:
                enabled: false
            kube-state-metrics:
              enabled: false
            prometheus-node-exporter:
              enabled: false
```

The receiver takes writes from anything that reaches the server, without
authentication, so keep the module's route internal.

- A Prometheus somewhere else: set `prometheusEndpoint` on both parts, the URL
  ends in `/api/v1/write`.
- Not writing to a Mimir or Prometheus that is part of the platform: set
  `mimirModule: ""` or `prometheusModule: ""` on both parts.
- `cluster-metrics.rules` works with Mimir only. A Prometheus keeps its rules
  in its own configuration.

Loki works the same way, with `node-agent.logs.lokiEndpoint` (ending in
`/loki/api/v1/push`) and `node-agent.logs.lokiTenantID`.

Log streams carry the same `cluster` label as the metrics, set by the agent;
change it with `node-agent.logs.externalLabels`. The logs the `receiver` takes in
through Faro get it too, from `receiver.faro.externalLabels`.
