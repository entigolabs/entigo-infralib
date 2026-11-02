## Opinionated helm package for grafana ##

In addition to installing grafana with Helm it also created the needed IRSA with crossplane.



### Azure ###

Azure Monitor datasource (metrics, Log Analytics, Resource Graph) with workload identity: a WorkloadIdentity
(crossplane-azure) gives grafana Monitoring Reader on the agent resource group and the AKS node resource group.

### Example code ###

```
    modules:
      - name: grafana
        source: grafana

```
