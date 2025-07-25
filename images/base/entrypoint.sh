#!/bin/bash
#set -x

[ -z $TF_VAR_prefix ] && echo "TF_VAR_prefix must be set" && exit 1
[ -z "$AWS_REGION" -a -z "$GOOGLE_REGION" ] && echo "AWS_REGION or GOOGLE_REGION must be set" && exit 1
[ -z $COMMAND ] && echo "COMMAND must be set" && exit 1
[ -z $INFRALIB_BUCKET ] && echo "INFRALIB_BUCKET must be set" && exit 1

export TF_IN_AUTOMATION=1

running_jobs() {
    local still_running=""
    local status_changed=0
    
    for p in $PIDS; do
        pid=$(echo $p | cut -d"=" -f1)
        name=$(echo $p | cut -d"=" -f2)
        
        # Skip if we already know about this job's completion
        if [[ $COMPLETED == *$p* ]] || [[ $FAIL == *$p* ]]; then
            continue
        fi
        
        if kill -0 $pid 2>/dev/null; then
            still_running="$still_running\n~ $(basename $name) Running"
        else
            # Check if job completed successfully
            if wait $pid 2>/dev/null; then
                echo "✓ $(basename $name) Done"
                COMPLETED="$COMPLETED $p"
            else
                echo "✗ $(basename $name) Failed"
                FAIL="$FAIL $p"
            fi
            status_changed=1
        fi
    done
    
    # Only show running jobs if the list has changed
    if [ "$still_running" != "$LAST_RUNNING" ]; then
        if [ ! -z "$still_running" ]; then
            echo -e "-------$still_running"
        fi
        LAST_RUNNING="$still_running"
    fi
}


if [ "$COMMAND" == "test" ]
then
  if [ ! -f go.mod ]
  then
    cd /common && go mod download -x && cd /app
    go mod init github.com/entigolabs/entigo-infralib
    go mod edit -require github.com/entigolabs/entigo-infralib-common@v0.0.0 -replace github.com/entigolabs/entigo-infralib-common=/common
    go mod tidy
  fi
  cd test && go test -timeout $ENTIGO_INFRALIB_TEST_TIMEOUT
  exit $?

#Prepare project filesystems for plan stages. When we plan then we need to get the current S3 bucket content
elif [ "$COMMAND" == "plan" -o "$COMMAND" == "plan-destroy" -o "$COMMAND" == "argocd-plan"  -o "$COMMAND" == "argocd-plan-destroy" ]
then
  echo "Need to copy project files from bucket $INFRALIB_BUCKET"
  if [ "$TERRAFORM_CACHE" != "true" ]
  then
    echo "Excluding .terraform cache."
    AWS_S3_EXCLUDE_TERRAFORM=(--exclude "*.terraform/*")
    GOOGLE_S3_EXCLUDE_TERRAFORM=(-x "\.terraform/.*")
  else
    AWS_S3_EXCLUDE_TERRAFORM=()
    GOOGLE_S3_EXCLUDE_TERRAFORM=()
  fi
  
  if [ ! -z "$GOOGLE_REGION" ]
  then
    mkdir -p /project/steps/$TF_VAR_prefix
    gsutil -m -q rsync -r ${GOOGLE_S3_EXCLUDE_TERRAFORM[@]} gs://${INFRALIB_BUCKET}/steps/$TF_VAR_prefix /project/steps/$TF_VAR_prefix
    cd /project
  else
    cd $CODEBUILD_SRC_DIR
    aws s3 cp s3://${INFRALIB_BUCKET}/steps/$TF_VAR_prefix ./steps/$TF_VAR_prefix --recursive --no-progress --quiet ${AWS_S3_EXCLUDE_TERRAFORM[@]}
  fi

  if [ ! -d "steps/$TF_VAR_prefix" ]
  then
    find .
    echo "Unable to find path "steps/$TF_VAR_prefix""
    exit 5
  fi
  cd "steps/$TF_VAR_prefix"
#Prepare project filesystems for apply stages. When we apply then we need to get the tar artifact.
elif [ "$COMMAND" == "apply" -o "$COMMAND" == "apply-destroy" -o "$COMMAND" == "argocd-apply" -o "$COMMAND" == "argocd-apply-destroy" ]
then
  if [ ! -z "$GOOGLE_REGION" ]
  then
    gsutil -m -q cp gs://${INFRALIB_BUCKET}/$TF_VAR_prefix-tf.tar.gz /project/tf.tar.gz 
    if [ $? -ne 0 ]
    then
      echo "Unable to find artifacts from plan stage! gs://${INFRALIB_BUCKET}/$TF_VAR_prefix-tf.tar.gz"
      exit 4
    fi
    cd /project/ && tar -xzf tf.tar.gz
  else
    if [ ! -f $CODEBUILD_SRC_DIR_Plan/tf.tar.gz ]
    then
      echo "Unable to find artifacts from plan stage! $CODEBUILD_SRC_DIR_Plan/tf.tar.gz"
      exit 4
    fi
    tar -xzf $CODEBUILD_SRC_DIR_Plan/tf.tar.gz
  fi
  cd "steps/$TF_VAR_prefix"
fi



#Prepare and check the environment for terraform (common for plan and apply)
if [ "$COMMAND" == "plan" -o "$COMMAND" == "plan-destroy" -o "$COMMAND" == "apply" -o "$COMMAND" == "apply-destroy" ]
then
  #Authenticate git repos if any.
  if [ -f /usr/bin/gitlogin.sh ]
  then
    /usr/bin/gitlogin.sh
  fi
  /usr/bin/ca-certificates.sh
  
  cat backend.conf
  if [ $? -ne 0 ]
  then
    echo "Unable to find backend.conf file"
    exit 100
  fi
  terraform init -input=false -backend-config=backend.conf
  if [ $? -ne 0 ]
  then
    echo "Terraform init failed."
    exit 14
  fi

#Prepare and check the environment for Kubernetes (common for plan and apply)
elif [ "$COMMAND" == "argocd-plan" -o "$COMMAND" == "argocd-apply" -o "$COMMAND" == "argocd-plan-destroy" -o "$COMMAND" == "argocd-apply-destroy" ]
then
  /usr/bin/ca-certificates.sh
  echo "COMMAND $COMMAND, cluster $KUBERNETES_CLUSTER_NAME region $AWS_REGION"
  if [ ! -z "$GOOGLE_REGION" ]
  then
    gcloud container clusters get-credentials $KUBERNETES_CLUSTER_NAME --region $GOOGLE_REGION --project $GOOGLE_PROJECT
    export PROVIDER="google"
    export ARGOCD_HOSTNAME=$(kubectl get httproute -n ${ARGOCD_NAMESPACE} -o jsonpath='{.items[*].spec.hostnames[*]}')
  else
    aws eks update-kubeconfig --name $KUBERNETES_CLUSTER_NAME --region $AWS_REGION
    export PROVIDER="aws"
    export ARGOCD_HOSTNAME=$(kubectl get ingress -n ${ARGOCD_NAMESPACE} -l app.kubernetes.io/component=server -o jsonpath='{.items[*].spec.rules[*].host}')
  fi
  export ARGOCD_AUTH_TOKEN=$(kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-infralib-token -o jsonpath="{.data.token}" | base64 -d)
  export USE_ARGOCD_CLI="false"
  if [ "$ARGOCD_AUTH_TOKEN" != "" -a "$ARGOCD_HOSTNAME" != "" ]
  then
    TES_CONNECTION=$(argocd --server ${ARGOCD_HOSTNAME} --grpc-web app list)
    if [ $? -eq 0 ]
    then
      echo "Connected to ArgoCD successfully."
      export USE_ARGOCD_CLI="true"
    fi
  fi
fi

#ALL SPECIFIC COMMANDS HERE
#Plan terraform
if [ "$COMMAND" == "plan" ]
then
  terraform plan -no-color -out ${TF_VAR_prefix}.tf-plan -input=false
  if [ $? -ne 0 ]
  then
    echo "Failed to create TF plan!"
    exit 6
  fi
elif [ "$COMMAND" == "apply" ]
then
  if [ "$TERRAFORM_CACHE" == "true" ]
  then
    echo "Syncing .terraform back to bucket"
    if [ ! -z "$GOOGLE_REGION" ]
    then
      gsutil -m -q rsync -d -r .terraform gs://${INFRALIB_BUCKET}/steps/$TF_VAR_prefix/.terraform
    else
      aws s3 sync .terraform s3://${INFRALIB_BUCKET}/steps/$TF_VAR_prefix/.terraform --no-progress --quiet --delete
    fi
  fi
  terraform apply -no-color -input=false ${TF_VAR_prefix}.tf-plan
  if [ $? -ne 0 ]
  then
    echo "Apply failed!"
    exit 11
  fi
  terraform output -json > terraform-output.json
  if [ $? -ne 0 ]
  then
    echo "Output failed!"
    exit 12
  fi
  if [ ! -z "$GOOGLE_REGION" ]
  then
    gsutil -m -q cp terraform-output.json gs://${INFRALIB_BUCKET}/$TF_VAR_prefix/terraform-output.json
  else
    aws s3 cp terraform-output.json s3://${INFRALIB_BUCKET}/$TF_VAR_prefix/terraform-output.json --no-progress --quiet
  fi
elif [ "$COMMAND" == "plan-destroy" ]
then
  terraform plan -destroy -no-color -out ${TF_VAR_prefix}.tf-plan-destroy -input=false
  if [ $? -ne 0 ]
  then
    echo "Failed to create TF destroy plan!"
    exit 6
  fi

elif [ "$COMMAND" == "apply-destroy" ]
then
  terraform apply -no-color -input=false ${TF_VAR_prefix}.tf-plan-destroy
  if [ $? -ne 0 ]
  then
    echo "Apply destroy failed!"
    exit 11
  fi
elif [ "$COMMAND" == "argocd-plan" ]
then
  HELM_BOOTSTAP="false"
  #When we first run then argocd is not yet installed and we can not use Application objects without installing it.
  if [ "$ARGOCD_HOSTNAME" == "" ]
  then
    #Authenticate git repos if any.
    if [ -f /usr/bin/gitlogin.sh ]
    then
      /usr/bin/gitlogin.sh
    fi
    echo "Detecting ArgoCD modules."
    for app_file in ./*.yaml
    do
      if yq -r '.spec.sources[0].path' $app_file | grep -q "modules/k8s/argocd"
      then
        echo "Found $app_file, installing using helm."
        app=`yq -r '.metadata.name' $app_file`
        yq -r '.spec.sources[0].helm.values' $app_file > values-$app.yaml
        namespace=`yq -r '.spec.destination.namespace' $app_file`
        version=`yq -r '.spec.sources[0].targetRevision' $app_file`
        repo=`yq -r '.spec.sources[0].repoURL' $app_file`
        path=`yq -r '.spec.sources[0].path' $app_file`
        git clone --depth 1 --single-branch --branch $version $repo git-$app
        #Create bootstrap value file that is only used first time ArgoCD is created.
        if compgen -A variable | grep -q "^GIT_AUTH_SOURCE_"
        then
          echo "
argocd:
  configs:
    repositories:" > git-$app/$path/extra_repos.yaml

          for var in "${!GIT_AUTH_SOURCE_@}"; do
            NAME=$(echo $var | sed 's/GIT_AUTH_SOURCE_//g')
            SOURCE="$(echo ${!var})"
            PASSWORD="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_PASSWORD/g')"
            USERNAME="$(echo $var | sed 's/GIT_AUTH_SOURCE/GIT_AUTH_USERNAME/g')"
            echo "      ${NAME}:
        url: ${SOURCE}.git
        name: ${NAME}
        password: ${!PASSWORD}
        username: ${!USERNAME}" >> git-$app/$path/extra_repos.yaml
          done
        else
            touch git-$app/$path/extra_repos.yaml
        fi
        
        helm upgrade --create-namespace --install -n $namespace -f git-$app/$path/values.yaml -f git-$app/$path/values-${PROVIDER}.yaml -f values-$app.yaml -f git-$app/$path/extra_repos.yaml --set-string 'argocd.configs.cm.admin\.enabled=true' --set argocd-apps.enabled=false $app git-$app/$path
        rm -rf values-$app.yaml git-$app
        HELM_BOOTSTAP="true"
      fi
    done
    if [ "$PROVIDER" == "google" ]
    then
      export ARGOCD_HOSTNAME=$(kubectl get httproute -n ${ARGOCD_NAMESPACE} -o jsonpath='{.items[*].spec.hostnames[*]}')
    else
      export ARGOCD_HOSTNAME=$(kubectl get ingress -n ${ARGOCD_NAMESPACE} -l app.kubernetes.io/component=server -o jsonpath='{.items[*].spec.rules[*].host}')
    fi
  fi

  if [ "$ARGOCD_HOSTNAME" == "" ]
  then
    echo "Unable to get ArgoCD hostname. Check ArgoCD installation."
    exit 25
  fi
  
  rm -f *.sync *.log
  PIDS=""
  for app_file in ./*.yaml
  do
      argocd-apps-plan.sh $app_file > $app_file.log 2>&1 &
      PIDS="$PIDS $!=$app_file"
  done

  FAIL=""
  COMPLETED=""
  LAST_RUNNING=""
  while true; do
      sleep 2
      running_jobs
      total_done=$(echo "$COMPLETED $FAIL" | wc -w)
      total_jobs=$(echo "$PIDS" | wc -w)
      
      if [ $total_done -eq $total_jobs ]; then
          break
      fi
  done

  for p in $COMPLETED; do
      name=$(echo $p | cut -d"=" -f2)
      cat $name.log
  done
  
  for p in $FAIL; do
      name=$(echo $p | cut -d"=" -f2)
      cat $name.log
  done
  
  ADD=`cat ./*.log | grep "^Status " | grep -ve"Status: Synced" | grep -ve "Missing:0" | wc -l`
  CHANGE=`cat ./*.log | grep "^Status " | grep -ve"Status: Synced" | grep -ve "Changed:0" | wc -l`
  DESTROY=`cat ./*.log | grep "^Status " | grep -ve"Status: Synced" | grep -ve "RequiredPruning:0" | wc -l`
  
  #Prevent agent from confirming first bootstrap when ArgoCD's own application will always show changes since it is already bootstrapped with Helm.
  if [ $HELM_BOOTSTAP == "true"  -a $CHANGE -gt 0 ]
  then
    CHANGE=$((CHANGE - 1))
  fi

  echo "ArgoCD Applications: ${ADD} to add, ${CHANGE} to change, ${DESTROY} to destroy."

  rm -f *.log
  
  if [ ! -z "$FAIL" ]; then
      echo "Failed jobs were:"
      for p in $FAIL; do
          echo "  - $(basename $(echo $p | cut -d"=" -f2))"
      done
      echo "Plan ArgoCD failed!"
      exit 21
  fi

elif [ "$COMMAND" == "argocd-apply" ]
then

  if [ "$ARGOCD_HOSTNAME" == "" ]
  then
    echo "Unable to get ArgoCD hostname."
    exit 25
  fi
  
  PIDS=""
  for app_file in ./*.yaml
  do
      argocd-apps-apply.sh $app_file > $app_file.log 2>&1 &
      PIDS="$PIDS $!=$app_file"
  done

  FAIL=""
  COMPLETED=""
  LAST_RUNNING=""
  while true; do
      sleep 2
      running_jobs
      total_done=$(echo "$COMPLETED $FAIL" | wc -w)
      total_jobs=$(echo "$PIDS" | wc -w)
      
      if [ $total_done -eq $total_jobs ]; then
          break
      fi
  done
  
  #Try the failed apps second time.
  PIDS=""
  for p in $FAIL; do
      name=$(echo $p | cut -d"=" -f2)
      argocd-apps-apply.sh $name > $name.log 2>&1 &
      PIDS="$PIDS $!=$name"
  done
  
  if [ ! -z "$FAIL" ]; then
      echo "Retry Failed jobs:"
      for p in $FAIL; do
          echo "  - $(basename $(echo $p | cut -d"=" -f2))"
      done
      FAIL=""
      COMPLETED=""
      LAST_RUNNING=""
      while true; do
          sleep 2
          running_jobs
          total_done=$(echo "$COMPLETED $FAIL" | wc -w)
          total_jobs=$(echo "$PIDS" | wc -w)
          if [ $total_done -eq $total_jobs ]; then
              break
          fi
      done
  fi

  if [ ! -z "$FAIL" ]; then
      for p in $FAIL; do
          name=$(echo $p | cut -d"=" -f2)
          echo "#######################################"
          echo "ERROR LOG FOR $name"
          cat $name.log
      done
  
      echo "Failed jobs were:"
      for p in $FAIL; do
          echo "  - $(basename $(echo $p | cut -d"=" -f2))"
      done
      echo "Apply ArgoCD failed!"
      exit 21
  fi
  rm -f *.log

elif [ "$COMMAND" == "argocd-plan-destroy" ]
then
  false
  if [ $? -ne 0 ]
  then
    echo "Plan ArgoCD destroy failed!"
    exit 22
  fi
elif [ "$COMMAND" == "argocd-apply-destroy" ]
then
  false
  if [ $? -ne 0 ]
  then
    echo "Apply ArgoCD destroy failed!"
    exit 23
  fi
else
  echo "Unknown command: $COMMAND"
  exit 1
fi 


#Compress artifacts created in plan stage that will be used in apply stage.
if [ "$COMMAND" == "argocd-plan-destroy" -o "$COMMAND" == "argocd-plan" -o "$COMMAND" == "plan-destroy" -o "$COMMAND" == "plan" ]
then
  cd ../..
  tar -czf tf.tar.gz "steps/$TF_VAR_prefix"
  if [ ! -z "$GOOGLE_REGION" ]
  then
    echo "Copy plan to Google S3"
    gsutil -m -q cp tf.tar.gz gs://${INFRALIB_BUCKET}/$TF_VAR_prefix-tf.tar.gz
  fi

fi
