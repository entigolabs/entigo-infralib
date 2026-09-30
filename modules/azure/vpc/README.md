## Opinionated module for Azure VNet creation ##

Creates a VNet with private, public, intra and database subnets, a delegated subnet for Application Gateway for
Containers (AGC), a delegated subnet for AKS API Server VNet Integration, a delegated pipeline subnet for Container
Apps (the agent jobs) and a NAT gateway for the private and pipeline subnets.

### Automatic subnet calculations ###
If only `vpc_cidr` is given, the subnets are calculated from it:

* private: the first half of the VNet
* public, intra, database: the first three quarters of the second half
* agc: the first /24 of the last quarter (AGC with Azure CNI Overlay needs exactly a /24)
* apiserver: the /28 after the AGC /24
* pipeline: the /27 after the API server /28 for a Container Apps workload profiles environment: the agent jobs of
  steps with `vpc: attach` (`pipeline_subnets` is the agent's default subnet) and other Container Apps in the same
  environment. One environment per subnet, delegated to `Microsoft.App/environments`, minimum /27 = 18 usable IPs (up
  to 90 Consumption replicas), the size can't change once an environment uses it.

The minimum VNet size for the automatic split is /20. For smaller VNets set `agc_subnets`, `apiserver_subnets` and
`pipeline_subnets` explicitly (or `[]` without AKS or Container Apps).

| vpc_cidr | private | public, intra, database | agc | apiserver | pipeline |
|---|---|---|---|---|---|
| /16 | /17 | /19 each | /24 | /28 | /27 |
| /18 | /19 | /21 each | /24 | /28 | /27 |
| /20 | /21 | /23 each | /24 | /28 | /27 |

Any subnet list can be set explicitly. The split does not shift for explicit lists, so overlapping subnets are
rejected at plan time.

Default outbound access is off on all subnets: the private and pipeline subnets egress through the NAT gateway
(`nat_static_ip_count` public IPs, about 64k SNAT ports each), intra and database subnets have no internet access and
public subnets only through a resource's own public IP or load balancer. Subnet names default to
`<prefix>-<type>-<n>`, override them with `*_subnet_names`. The private subnets have service endpoints for Storage and
Key Vault: these services then see the subnet's private IP instead of the NAT IP, so IP firewall rules need VNet
rules instead.

### Example code ###
```
    modules:
      - name: vpc
        source: azure/vpc
        inputs:
          vpc_cidr: "10.156.0.0/16"
```
Will result in networks:

private-0 ( 10.156.0.0/17 )

public-0 ( 10.156.128.0/19 )

intra-0 ( 10.156.160.0/19 )

database-0 ( 10.156.192.0/19 )

agc-0 ( 10.156.224.0/24 )

apiserver-0 ( 10.156.225.0/28 )

pipeline-0 ( 10.156.225.32/27 )
