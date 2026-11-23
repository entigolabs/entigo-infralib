# kube-state-metrics

kube-state-metrics on its own: the `kube_*` metrics about the objects of the
cluster (pods, deployments, nodes, ...), which most Kubernetes dashboards are
built on.

Only install it on a cluster that does **not** run the `prometheus` module.
The `prometheus` module brings its own kube-state-metrics, and a second one
makes every `kube_*` series exist twice.

It is the companion of the `alloy` module with metrics enabled: Alloy's
`cluster-metrics` scrapes it and writes to Mimir, see
[alloy](../alloy/README.md#metrics).

```yaml
    modules:
      - name: kube-state-metrics
        source: kube-state-metrics
      - name: alloy
        source: alloy
        inputs:
          node-agent:
            metrics:
              enabled: true
          cluster-metrics:
            enabled: true
```

The Deployment and its Service are always called `kube-state-metrics`,
whatever the module is named. Node and namespace labels that the dashboards
and the Zone model use (instance type, architecture, availability zone,
`tenancy.entigo.com/zone`) are exported as labels of `kube_node_labels` and
`kube_namespace_labels`; extend the list with
`kube-state-metrics.metricLabelsAllowlist`.
