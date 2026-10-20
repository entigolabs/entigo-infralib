# cilium

Enforces Kubernetes NetworkPolicies on OKE, whose VCN-native CNI accepts them but does not enforce
them. Cilium joins the OCI CNI's chain (`cni.chainingTarget: oci`): the OCI CNI keeps pod IP
allocation and routing, and Cilium enforces policy on each pod. The settings follow Cilium's
[OKE chaining guide](https://docs.cilium.io/en/stable/installation/cni-chaining-oracle-oke/).

## Pods started before Cilium are not policed

Cilium attaches to a pod when the pod starts. Pods already running when Cilium is installed keep
working but are not policed until they restart. The module syncs at wave 0, ahead of every other
application, so on a new cluster only system pods and Argo CD, which the agent installs before any
application syncs, predate it.

## Allow the API server by its endpoint, not its Service IP

Cilium checks a connection's destination after Service translation, so an egress rule on the
`kubernetes` Service IP (`10.96.0.1:443`) never matches. Allow the cluster's private endpoint
instead: the `oracle/oke` output `private_endpoint_ip`, port 6443. `modules/k8s/prometheus`
does this for `kube-state-metrics`.

## Removing the module breaks new pods until the nodes are cleaned

Uninstalling leaves `/etc/cni/net.d/05-cilium.conflist` on every node. It sorts before the OCI
CNI's `10-oci.conflist`, so new pods stay in `ContainerCreating` until the file is deleted on
each node or the nodes are replaced.
