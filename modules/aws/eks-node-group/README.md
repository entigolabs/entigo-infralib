## Oppinionated module for eks managed node group creation ##


Oppinionated version of this https://github.com/terraform-aws-modules/terraform-aws-eks/tree/master/modules/eks-managed-node-group


__node_repair_config__ Defaults to null (node auto repxhealth issues, it is installed by the aws/eks module when its node_repair_config is enabled or enable_node_monitoring_agent is true.

__node_repair_wait_time_mins__ Defaults to 20, minutes before repairing a node for all repair triggering node conditions (10-120, multiple of 10). Ignored when node_repair_config sets node_repair_config_overrides.

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
