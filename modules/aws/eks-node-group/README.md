## Oppinionated module for eks managed node group creation ##


Oppinionated version of this https://github.com/terraform-aws-modules/terraform-aws-eks/tree/master/modules/eks-managed-node-group


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
