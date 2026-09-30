## Opinionated module for an extra AKS node pool ##

Adds a node pool to an azure/aks cluster with the [AVM agentpool module](https://github.com/Azure/terraform-azurerm-avm-res-containerservice-managedcluster/tree/main/modules/agentpool),
like aws/eks-node-group and google/gke-node-pool. Cluster, version and node subnet come from azure/aks.

* Pool name: AKS allows 1-12 lowercase letters and digits, so instead of the whole prefix the pool is named after the
  agent module name (`name`, the prefix without `<config prefix>-<step>-`), with other characters removed and cut to 12:
  module `spot` => `spot`, `node-pool-spot` => `nodepoolspot`. Module names must differ within their first 12 letters
  and digits.
* Autoscaling between `min_size` and `max_size`. `spot_nodes` for spot VMs: AKS taints and labels them itself
  (`kubernetes.azure.com/scalesetpriority=spot:NoSchedule` and `kubernetes.azure.com/scalesetpriority: spot`), so
  workloads for spot nodes tolerate and select that; no surge upgrades. `taints` as `key=value:Effect` strings.
* Zones: by default every zone where `instance_type` is available in the region (Compute SKU list), set
  `availability_zones` to override (`[]` = no zones). Zones are fixed when the pool is created.
* Encryption, kubelet identity (AcrPull) and the bootstrap artifact cache come from the cluster: the disk encryption
  set of azure/aks (kms data key) also applies to this pool.
* Node image updates follow the cluster's `node_os_upgrade_channel`; `kubernetes_version` upgrades the pool.

### Example code ###
```
    modules:
      - name: spot
        source: azure/aks-node-pool
        inputs:
          spot_nodes: true
          min_size: 0
          max_size: 3
```
