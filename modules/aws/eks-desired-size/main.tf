resource "aws_ssm_parameter" "eks_main_min_size" {
  name  = "/entigo-infralib/${local.hname}/eks_main_min_size"
  type  = "String"
  value = var.eks_main_min_size
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_mainarm_min_size" {
  name  = "/entigo-infralib/${local.hname}/eks_mainarm_min_size"
  type  = "String"
  value = var.eks_mainarm_min_size
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_tools_min_size" {
  name  = "/entigo-infralib/${local.hname}/eks_tools_min_size"
  type  = "String"
  value = var.eks_tools_min_size
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_mon_min_size" {
  name  = "/entigo-infralib/${local.hname}/eks_mon_min_size"
  type  = "String"
  value = var.eks_mon_min_size
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_spot_min_size" {
  name  = "/entigo-infralib/${local.hname}/eks_spot_min_size"
  type  = "String"
  value = var.eks_spot_min_size
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "eks_db_min_size" {
  name  = "/entigo-infralib/${local.hname}/eks_db_min_size"
  type  = "String"
  value = var.eks_db_min_size
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "null_resource" "update_desired_size" {

  triggers = {
    main = var.eks_main_min_size
    mainarm = var.eks_mainarm_min_size
    tools = var.eks_tools_min_size
    mon = var.eks_mon_min_size
    spot = var.eks_spot_min_size
    db = var.eks_db_min_size
    always_run = timestamp() # causes Terraform to always run this module, even if nothing changes. Needed for testing.
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]

    environment = {
      eks_main_min_size = local.eks_min_size_map["main"]
      eks_mainarm_min_size  = local.eks_min_size_map["mainarm"]
      eks_tools_min_size = local.eks_min_size_map["tools"]
      eks_mon_min_size = local.eks_min_size_map["mon"]
      eks_spot_min_size = local.eks_min_size_map["spot"]
      eks_db_min_size = local.eks_min_size_map["db"]
    }

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

        # Set the new minimum size based on the short name
        if [ "$node_group_short_name" == "main" ]; then
          new_min_size=$eks_main_min_size
        elif [ "$node_group_short_name" == "mainarm" ]; then
          new_min_size=$eks_mainarm_min_size
        elif [ "$node_group_short_name" == "tools" ]; then
          new_min_size=$eks_tools_min_size
        elif [ "$node_group_short_name" == "mon" ]; then
          new_min_size=$eks_mon_min_size
        elif [ "$node_group_short_name" == "spot" ]; then
          new_min_size=$eks_spot_min_size
        elif [ "$node_group_short_name" == "db" ]; then
          new_min_size=$eks_db_min_size
        else
          echo "Unknown node group short name: $node_group_short_name"
          continue
        fi

        current_desired_size=$(printf "%d" "$current_desired_size")

        echo "New min size: $new_min_size"
        echo "Current desired size: $current_desired_size"

        # Check if current desired size is less than new min size, if true then update node group desired size to new min size
        if [ $current_desired_size -lt $new_min_size ]; then
          aws eks update-nodegroup-config --cluster-name ${var.cluster_name} --nodegroup-name $nodegroup --scaling-config desiredSize=$new_min_size
          echo "Updated node group $nodegroup to new min size: $new_min_size"
        else
          echo "Node group $nodegroup already at min size: $new_min_size". No update needed
        fi

        # Check if node group is in ACTIVE state, if not then sleep for 5 seconds and check again
        while [ $(aws eks describe-nodegroup --cluster-name ${var.cluster_name} --nodegroup-name $nodegroup --query "nodegroup.status" --output text) != "ACTIVE" ]; do
          sleep 5
        done

      done

    EOT
  }
}

