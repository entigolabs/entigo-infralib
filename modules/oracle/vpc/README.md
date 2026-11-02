## Opinionated module for an OCI VCN and its subnets ##

Creates the subnets a deployment runs in - `private`, `pod`, `public`, `intra` and `database` -
carved from `vpc_cidr` unless listed explicitly. By default it also creates the VCN itself, an
internet, NAT and service gateway, and one route table per tier: `public` to the internet
gateway, `private` and `pod` to the NAT and service gateways, `intra` and `database` to the
service gateway only.

### With `vcn_id`, the module creates only subnets ###

Setting `vcn_id` puts the subnets into a VCN that already exists, typically one a central network
team owns in its own compartment and routes through a hub. The module then reads the VCN and:

- takes `vpc_cidr` from the VCN's first CIDR block, so the subnet layout follows the VCN;
- creates the subnets in the VCN's compartment, or in `network_compartment_id` when set;
- creates no gateways and no route tables, so every subnet uses the VCN's default route table;
- prohibits public IPs in the `public` tier, since without an internet gateway they carry no
  traffic.

Every one of those choices can be turned back on: `enable_internet_gateway`,
`enable_nat_gateway`, `enable_service_gateway`, `create_route_tables` and
`public_subnets_allow_public_ip`. A gateway requires `create_route_tables`, and the internet
gateway also requires `public_subnets_allow_public_ip`; preconditions reject the combinations
where nothing would route to the gateway. In both modes the subnets use the VCN's default
security list and DHCP options.

When the VCN cannot be read, the plan fails with `VCN <id> returned no CIDR blocks`. Check
`vcn_id`, and that the identity running the plan can read VCNs in the VCN's compartment.

### The network compartment needs a policy ###

With the subnets in another compartment, three kinds of resource principal in the deployment's
compartment act on them, and the network compartment's owner has to allow each. Resource
principals cannot be members of a group, so the statements match them by
`request.principal.type` and by the compartment they run in.

Replace `<network-compartment-ocid>` with the VCN's compartment and `<workload-compartment-ocid>`
with the compartment the deployment runs in:

```
Allow any-user to read   virtual-network-family in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>' }
Allow any-user to read   dns                    in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>' }
Allow any-user to manage subnets                in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>' }
Allow any-user to manage vcns                   in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>', any { request.operation = 'CreateSubnet', request.operation = 'UpdateSubnet', request.operation = 'DeleteSubnet', request.operation = 'CreateNetworkSecurityGroup', request.operation = 'DeleteNetworkSecurityGroup' } }
Allow any-user to manage route-tables           in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>', any { request.operation = 'CreateSubnet', request.operation = 'UpdateSubnet', request.operation = 'DeleteSubnet' } }
Allow any-user to manage security-lists         in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>', any { request.operation = 'CreateSubnet', request.operation = 'UpdateSubnet', request.operation = 'DeleteSubnet' } }
Allow any-user to manage dhcp-options           in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>', any { request.operation = 'CreateSubnet', request.operation = 'UpdateSubnet', request.operation = 'DeleteSubnet' } }
Allow any-user to use    vnics                  in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>' }
Allow any-user to use    private-ips            in compartment id <network-compartment-ocid> where all { request.principal.type = 'devopsbuildpipeline', request.principal.compartment.id = '<workload-compartment-ocid>' }

Allow any-user to use    private-ips            in compartment id <network-compartment-ocid> where all { request.principal.type = 'cluster', request.principal.compartment.id = '<workload-compartment-ocid>' }

Allow any-user to use    subnets                in compartment id <network-compartment-ocid> where all { request.principal.type = 'workload', request.principal.namespace = 'native-ingress-controller-system', request.principal.service_account = 'oci-native-ingress-controller', request.principal.compartment.id = '<workload-compartment-ocid>' }
Allow any-user to use    vnics                  in compartment id <network-compartment-ocid> where all { request.principal.type = 'workload', request.principal.namespace = 'native-ingress-controller-system', request.principal.service_account = 'oci-native-ingress-controller', request.principal.compartment.id = '<workload-compartment-ocid>' }
Allow any-user to use    private-ips            in compartment id <network-compartment-ocid> where all { request.principal.type = 'workload', request.principal.namespace = 'native-ingress-controller-system', request.principal.service_account = 'oci-native-ingress-controller', request.principal.compartment.id = '<workload-compartment-ocid>' }
```

**Build pipelines**, the first nine, run the Terraform steps:

- `read virtual-network-family` lets `oracle/oke` create the cluster. `CreateCluster` returns
  `404 NotAuthorizedOrNotFound` with only `read vcns` or `inspect vcns`.
- `read dns` lets `oracle/dns` find the VCN's resolver, whose default view holds the private zones.
- Creating a subnet, or one of `oracle/oke`'s network security groups, in a VCN is authorised
  against the VCN as well, and a subnet also against the route table, security list and DHCP
  options it uses. The `manage` statements on those resources are limited to exactly those
  operations, so they grant no way to change the VCN or its routing.
- `use vnics` and `use private-ips` let them place network interfaces in the subnets.

**The cluster**, the tenth, allocates pod IPs in the `pod` subnet.

**The ingress controller**, the last three, places the internal load balancer in the `private`
subnet. The statements name the namespace and service account of
`modules/k8s/oci-native-ingress-controller`: they follow its `deploymentNamespace` value and the
`oci-native-ingress-controller` service account in `templates/policy.yaml`.

The deployment also needs outbound access through the hub to the region's Oracle services, for
nodes to register with the cluster, and to the container registries its applications pull from.
