#!/bin/bash
#set -x

[ -z $TF_VAR_prefix ] && echo "TF_VAR_prefix must be set" && exit 1
[ -z "$AWS_REGION" -a -z "$GOOGLE_REGION" ] && echo "AWS_REGION or GOOGLE_REGION must be set" && exit 1
[ -z $COMMAND ] && echo "COMMAND must be set" && exit 1
[ -z $INFRALIB_BUCKET ] && echo "INFRALIB_BUCKET must be set" && exit 1

export TF_IN_AUTOMATION=1

#Prepare project filesystems for plan stages. When we plan then we need to get the current S3 bucket content
if [ "$COMMAND" == "plan" -o "$COMMAND" == "plan-destroy" -o "$COMMAND" == "argocd-plan"  -o "$COMMAND" == "argocd-plan-destroy" ]
then
  echo "Need to copy project files from bucket $INFRALIB_BUCKET"
  if [ ! -z "$GOOGLE_REGION" ]
  then
    mkdir /project/steps/
    gsutil -m -q cp -r gs://${INFRALIB_BUCKET}/steps/$TF_VAR_prefix /project/steps/
    cd /project
  else
    cd $CODEBUILD_SRC_DIR
    aws s3 cp s3://${INFRALIB_BUCKET}/steps/$TF_VAR_prefix ./steps/$TF_VAR_prefix --recursive --no-progress --quiet
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
  /usr/bin/gitlogin.sh
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
  echo "Syncing .terraform back to bucket"
  if [ ! -z "$GOOGLE_REGION" ]
  then
    gsutil -m -q rsync -d -r .terraform gs://${INFRALIB_BUCKET}/steps/$TF_VAR_prefix/.terraform
  else
    aws s3 sync .terraform s3://${INFRALIB_BUCKET}/steps/$TF_VAR_prefix/.terraform --no-progress --quiet --delete
  fi
  terraform apply -no-color -input=false ${TF_VAR_prefix}.tf-plan
  if [ $? -ne 0 ]
  then
    echo "Apply failed!"
    exit 11
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
  #When we first run then argocd is not yet installed and we can not use Application objects without installing it.
  if [ "$ARGOCD_HOSTNAME" == "" ]
  then
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
        helm upgrade --create-namespace --install -n $namespace -f git-$app/$path/values.yaml -f git-$app/$path/values-${PROVIDER}.yaml -f values-$app.yaml $app git-$app/$path
        rm -rf values-$app.yaml git-$app
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
  export ARGOCD_AUTH_TOKEN=`kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-infralib-token -o jsonpath="{.data.token}" | base64 -d`
  
  if [ "$ARGOCD_AUTH_TOKEN" == "" ]
  then
    echo "No infralib ArgoCD token found, probably it is first run. Trying to create token using admin credentials."
    ARGO_PASS=`kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d` 
    argocd login --password ${ARGO_PASS} --username admin ${ARGOCD_HOSTNAME} --grpc-web
    export ARGOCD_AUTH_TOKEN=`argocd account generate-token --account infralib`
    argocd logout ${ARGOCD_HOSTNAME}
    if [ "$ARGOCD_AUTH_TOKEN" != "" ]
    then
      kubectl create secret -n ${ARGOCD_NAMESPACE} generic argocd-infralib-token --from-literal=token=$ARGOCD_AUTH_TOKEN
    else
      echo "Failed to create ARGOCD_AUTH_TOKEN. This is normal initially when the ArgoCD ingress hostname is not resolving yet."
    fi
  fi
  rm -f *.sync *.log
  PIDS=""
  for app_file in ./*.yaml
  do
      argocd-apps-plan.sh $app_file > $app_file.log 2>&1 &
      PIDS="$PIDS $!"
  done

  FAIL=0
  for p in $PIDS; do
      wait $p || let "FAIL+=1"
  done

  for app_log_file in ./*.log
  do
    cat $app_log_file
  done

  if [ "$ARGOCD_AUTH_TOKEN" != "" ]
  then
    CHANGED=`cat ./*.log | grep "^Status " | grep -ve"Status: Synced" | grep "RequiredPruning: 0$" | wc -l`
    DESTROY=`cat ./*.log | grep "^Status " | grep -ve"Status: Synced" | grep -ve "RequiredPruning: 0$" | wc -l`
  else
    CHANGED=`cat ./*.log | grep '^application\.argoproj\.io/.*' | wc -l`
    DESTROY="0"
  fi
  echo "ArgoCD Applications: ${CHANGED} has changed objects, ${DESTROY} has RequiredPruning objects."

  rm -f *.log
  
  if [ "$FAIL" -ne 0 ]
  then
    echo "FAILED to plan $FAIL applications."
    echo "Plan ArgoCD failed!"
    exit 20
  fi
  
elif [ "$COMMAND" == "argocd-apply" ]
then

  if [ "$ARGOCD_HOSTNAME" == "" ]
  then
    echo "Unable to get ArgoCD hostname."
    exit 25
  fi
  export ARGOCD_AUTH_TOKEN=`kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-infralib-token -o jsonpath="{.data.token}" | base64 -d`

  
  PIDS=""
  for app_file in ./*.yaml
  do
      argocd-apps-apply.sh $app_file > $app_file.log 2>&1 &
      PIDS="$PIDS $!"
  done

  FAIL=0
  for p in $PIDS; do
      wait $p || let "FAIL+=1"
  done

  for app_log_file in ./*.log
  do
    cat $app_log_file
    rm $app_log_file
  done

  if [ "$FAIL" -ne 0 ]
  then
    echo "FAILED to apply $FAIL applications."
    echo "Apply ArgoCD failed!"
    exit 21
  fi

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
