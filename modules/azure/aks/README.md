## Opinionated module for AKS creation ##

AKS cluster with the [AVM managed cluster module](https://github.com/Azure/terraform-azurerm-avm-res-containerservice-managedcluster):
Azure CNI Overlay + Cilium, NAT gateway egress (vpc), Workload Identity + OIDC issuer, Entra ID auth with Azure RBAC
(no local admin account) and API Server VNet Integration in the vpc apiserver subnet.

* Node pools `main`, `mon` and `tools` like aws/eks and google/gke (labels `<pool>=true`, `mon=true:NoSchedule`).
  **`tools` is the AKS system pool** with the taint `CriticalAddonsOnly=true:NoSchedule` instead of `tools=true`:
  AKS system pods only tolerate that taint, so the tools charts tolerate it next to their `tools` toleration. The
  system pool can't be spot and can't be removed; `main` and `mon` are left out with `aks_<pool>_max_size: 0`.
  More pools: `aks_managed_node_groups_extra` or azure/aks-node-pool.
* Pools are spread over every zone where their VM size is available in the region (read from the Compute SKU list),
  `availability_zones` overrides it (`[]` = no zones). Zones, VM size and max pods are fixed per pool at creation.
* Non-spot pools upgrade with `max_surge` (default 10%), disks are `aks_<pool>_volume_type` (Managed or Ephemeral),
  nodes carry `created-by=entigo-infralib`. Cluster autoscaler settings: `auto_scaler_profile` (AKS defaults when
  null); the k8s cluster-autoscaler module is not used on Azure.
* Node resource group `<prefix>-nodes-<location>` (`node_resource_group_name`), created and managed by AKS, only
  settable at cluster creation.
* Version: `kubernetes_version` is a minor (N-1 of the latest GA, like the stable channel), AKS picks the latest
  patch at creation. No automatic upgrades by default: nodes get a new node image with every Kubernetes version
  upgrade or `az aks nodepool upgrade --node-image-only` (AKS can't pin a node image version like aws AMI releases).
  Automatic: `upgrade_channel: patch` and/or `node_os_upgrade_channel: NodeImage` (or `SecurityPatch`), both in the
  weekly `maintenance_window` (default Sunday 00:00 UTC, 4 hours).
* API server: private by default (`private_cluster_enabled`, like aws/eks and google/gke): the FQDN only resolves
  inside the VNet, so agent steps that use the cluster run in the vpc pipeline subnet (argocd steps attach by
  default) and people use the VPN. With
  `private_cluster_enabled: false` the public endpoint is limited to `api_server_authorized_ip_ranges` (`[]` = open).
* Access: the Terraform caller (the agent) and `admin_object_ids` get Azure Kubernetes Service RBAC Cluster Admin.
  The kubelet identity gets AcrPull on `acr_id` and `kubelet_additional_role_assignments`.
* `agc_subnet_ids` (vpc `agc_subnets`) creates the workload identity of the ALB controller (k8s azure-gateway, one
  per cluster, service account `azure-alb-system/alb-controller-sa`): Reader + AppGw for Containers Configuration
  Manager on the node resource group, Network Contributor on every AGC subnet (one Application Gateway for Containers
  per /24). Output `alb_controller_client_id`.
* Encryption: when the kms module exists, its data key (`disk_encryption_key_id`) encrypts the OS disks of all pools
  (incl. azure/aks-node-pool) and by default the PVC disks through a disk encryption set, **only at cluster creation**
  (adding kms later fails the apply, the cluster isn't replaced). No etcd encryption with the kms config key yet:
  AKS KMS data encryption is still preview.
* `bootstrap_cache_enabled`: nodes pull AKS system images through acr-proxy (Premium, private endpoint,
  `aks_bootstrap_cache_rule`) instead of MCR.
* Control plane logs to Log Analytics with `control_plane_logs_enabled`, off by default (billed per ingested GB),
  encrypted with Microsoft managed keys (customer managed keys need a dedicated Log Analytics cluster).
* Egress uses the vpc NAT gateway (`outbound_type: userAssignedNATGateway`): a vpc without NAT gateway fails at cluster
  creation.

### Example code ###
```
    modules:
      - name: aks
        source: azure/aks
        inputs:
          aks_main_min_size: 2
          aks_main_max_size: 4
```
