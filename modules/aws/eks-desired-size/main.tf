resource "aws_ssm_parameter" "eks_min_size" {
  for_each = local.eks_min_size_map

  name  = "/entigo-infralib/${local.hname}/${each.key}"
  type  = "String"
  value = each.value

  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "null_resource" "update_desired_size" {

  triggers = {
    eks_min_size_map = jsonencode([
      for key in sort(keys(local.eks_min_size_map)) : {
        key   = key
        value = local.eks_min_size_map[key]
      }
    ])
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]

    environment = local.eks_min_size_map

    command = <<-EOT

      # Check if cluster exists
      aws eks describe-cluster --name ${var.cluster_name} > /dev/null 2>&1
      if [ $? -ne 0 ]
      then
        echo "Cluster ${var.cluster_name} does not exist"
        exit 0
      else
        echo "Cluster ${var.cluster_name} exists"
      fi

      # Get list of node groups
      nodegroups=$(aws eks list-nodegroups --cluster-name ${var.cluster_name} --query "nodegroups" --output text)
      
      # Loop through each node group
      for nodegroup in $nodegroups; do
        echo "Nodegroup: $nodegroup"

        # Check if node group is in ACTIVE state, if not then sleep for 5 seconds and check again
        while [ $(aws eks describe-nodegroup --cluster-name ${var.cluster_name} --nodegroup-name $nodegroup --query "nodegroup.status" --output text) != "ACTIVE" ]; do
          sleep 5
        done
        
        # Get the current desired size of the node group
        current_desired_size=$(aws eks describe-nodegroup --cluster-name ${var.cluster_name} --nodegroup-name $nodegroup --query "nodegroup.scalingConfig.desiredSize" --output text)

        # Get the short name of the node group (main, mainarm, tools, mon, spot, db)
        node_group_short_name=$(echo "$nodegroup" | awk -F'-' '{print $(NF-1)}')
        echo "Node group short name: $node_group_short_name"
        
        new_min_size=0

        min_size_variable_name="eks_$${node_group_short_name}_min_size"
        echo "min_size_variable_name: $min_size_variable_name"

        new_min_size=$${!min_size_variable_name}

        current_desired_size=$(printf "%d" "$current_desired_size")
        new_min_size=$(printf "%d" "$new_min_size")

        echo "New min size: $new_min_size"
        echo "Current desired size: $current_desired_size"

        # Check if current desired size is less than new min size, if true then update node group desired size to new min size
        if [ $current_desired_size -lt $new_min_size ]; then
          aws eks update-nodegroup-config --cluster-name ${var.cluster_name} --nodegroup-name $nodegroup --scaling-config desiredSize=$new_min_size
          echo "Updated node group $nodegroup to new min size: $new_min_size"
        else
          echo "Node group $nodegroup already at min size: $new_min_size". No update needed.
        fi

        # Check if node group is in ACTIVE state, if not then sleep for 5 seconds and check again
        while [ $(aws eks describe-nodegroup --cluster-name ${var.cluster_name} --nodegroup-name $nodegroup --query "nodegroup.status" --output text) != "ACTIVE" ]; do
          sleep 5
        done

      done

    EOT
  }
}

