## Oppinionated module for eks managed node group creation ##


Oppinionated version of this https://github.com/terraform-aws-modules/terraform-aws-eks/tree/master/modules/eks-managed-node-group


__node_repair_config__ Defaults to null (node auto repair disabled). EKS managed node group auto repair configuration, enable with { enabled = true }. See [aws_eks_node_group node_repair_config](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_node_group#node_repair_config) for other options. Install the node monitoring agent addon with the aws/eks module input enable_node_monitoring_agent to detect more node health issues.

### Example code ###

```
        - name: mypool
          source: aws/eks-node-group
          inputs:
            min_size: 1
            desired_size: 1
            max_size: 1
            volume_size: 20
            instance_types: |
              ["t3.large"]
            taints: |
                [
                  {
                    key = "mypool"
                    value = "true"
                    effect = "NO_SCHEDULE"
                  }
                ]
            labels: |
                {
                  kcp = "mypool"
                }

```
