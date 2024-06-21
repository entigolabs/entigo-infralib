#!/bin/bash
set -x

[ -z $TF_VAR_prefix ] && echo "TF_VAR_prefix must be set" && exit 1
[ -z "$AWS_REGION" -a -z "$GOOGLE_REGION" ] && echo "AWS_REGION or GOOGLE_REGION must be set" && exit 1
[ -z $COMMAND ] && echo "COMMAND must be set" && exit 1
[ -z $WORKSPACE ] && echo "WORKSPACE must be set" && exit 1
export TF_IN_AUTOMATION=1

if [ ! -z "$GOOGLE_REGION" ]
then
  sleep 1
  cp -r $CODEBUILD_SRC_DIR/* /project/
  cd /project
else
  cd $CODEBUILD_SRC_DIR
fi

if [ -d ".git" ]
then
rm -rf .git
fi

if [ "$COMMAND" == "plan" -o "$COMMAND" == "plan-destroy" -o "$COMMAND" == "argocd-plan" -o "$COMMAND" == "argocd-apply" -o "$COMMAND" == "argocd-plan-destroy" -o "$COMMAND" == "argocd-apply-destroy" ]
then

  if [ ! -d "$TF_VAR_prefix/$WORKSPACE" ]
  then
    find .
    echo "Unable to find path "$TF_VAR_prefix/$WORKSPACE""
    exit 5
  fi
  cd "$TF_VAR_prefix/$WORKSPACE"

elif [ "$COMMAND" == "apply" -o "$COMMAND" == "apply-destroy" ]
then
  if [ ! -z "$GOOGLE_REGION" ]
  then
    if [ ! -f $TF_VAR_prefix-$WORKSPACE-tf.tar.gz ]
    then
      echo "Unable to find artifacts from plan stage! $TF_VAR_prefix-$WORKSPACE-tf.tar.gz"
      exit 4
    fi
    tar -xvzf $TF_VAR_prefix-$WORKSPACE-tf.tar.gz
  else
    if [ ! -f $CODEBUILD_SRC_DIR_Plan/tf.tar.gz ]
    then
      echo "Unable to find artifacts from plan stage! $CODEBUILD_SRC_DIR_Plan/tf.tar.gz"
      exit 4
    fi
    tar -xvzf $CODEBUILD_SRC_DIR_Plan/tf.tar.gz
  fi
  cd "$TF_VAR_prefix/$WORKSPACE"
fi

if [ "$COMMAND" == "plan" -o "$COMMAND" == "plan-destroy" -o "$COMMAND" == "apply" -o "$COMMAND" == "apply-destroy" ]
then
  /usr/bin/gitlogin.sh
  cat ../backend.conf
  terraform init -input=false -backend-config=../backend.conf
  if [ $? -ne 0 ]
  then
    echo "Terraform init failed."
    exit 14
  fi
  terraform workspace select $WORKSPACE -no-color || terraform workspace new $WORKSPACE -no-color && terraform workspace select $WORKSPACE -no-color && terraform init -input=false || exit 2
fi


if [ "$COMMAND" == "plan" ]
then
  terraform plan -no-color -out $WORKSPACE.tf-plan -input=false
  if [ $? -ne 0 ]
  then
    echo "Failed to create TF plan!"
    exit 6
  fi
  cd ../..
  tar -czf tf.tar.gz "$TF_VAR_prefix/$WORKSPACE"
  if [ ! -z "$GOOGLE_REGION" ]
  then
    echo "Copy plan to Google S3"
    env
    cp tf.tar.gz $CODEBUILD_SRC_DIR/$TF_VAR_prefix-$WORKSPACE-tf.tar.gz
  fi
elif [ "$COMMAND" == "apply" ]
then
  terraform apply -no-color -input=false $WORKSPACE.tf-plan
  if [ $? -ne 0 ]
  then
    echo "Apply failed!"
    exit 11
  fi
elif [ "$COMMAND" == "plan-destroy" ]
then
  terraform plan -destroy -no-color -out $WORKSPACE.tf-plan -input=false
  if [ $? -ne 0 ]
  then
    echo "Failed to create TF destroy plan!"
    exit 6
  fi
  cd ../..
  tar -czf tf.tar.gz "$TF_VAR_prefix/$WORKSPACE"
  if [ ! -z "$GOOGLE_REGION" ]
  then
    echo "Copy plan to Google S3"
    cp tf.tar.gz $CODEBUILD_SRC_DIR/$TF_VAR_prefix-$WORKSPACE-tf.tar.gz
  fi
elif [ "$COMMAND" == "apply-destroy" ]
then
  terraform apply -no-color -input=false $WORKSPACE.tf-plan
  if [ $? -ne 0 ]
  then
    echo "Apply destroy failed!"
    exit 11
  fi
elif [ "$COMMAND" == "argocd-plan" ]
then
  #gcloud auth activate-service-account --key-file=application_default_credentials.json
  #gcloud config set account $(gcloud auth list --filter=status:ACTIVE --format="value(account)")
  gcloud container clusters get-credentials $KUBERNETES_CLUSTER_NAME --region $GOOGLE_REGION --project $GOOGLE_PROJECT
  export ARGOCD_HOSTNAME=`kubectl get ingress -n argocd -l app.kubernetes.io/component=server  -o jsonpath='{.items[*].spec.rules[*].host}'`
  echo "ArgoCD hostname is $ARGOCD_HOSTNAME"
  
  export ARGO_TOKEN=`kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-infralib-token -o jsonpath="{.data.token}" | base64 -d`
  
  if [ "$ARGO_TOKEN" == "" ]
  then
    echo "No infralib argocd token found, probably it is first run. Trying to create token using admin credentials."
    ARGO_PASS=`kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d` 
    argocd login --password ${ARGO_PASS} --username admin ${ARGOCD_HOSTNAME} --grpc-web
    export ARGO_TOKEN=`argocd account generate-token --account infralib`
    argocd logout ${ARGOCD_HOSTNAME}
    kubectl create secret -n ${ARGOCD_NAMESPACE} generic argocd-infralib-token --from-literal=token=$ARGO_TOKEN
  fi
  
  find . -type f -name '*.yaml' | while read line
  do
    kubectl apply -n ${ARGOCD_NAMESPACE} -f $line
    if [ $? -ne 0 ]
    then
      echo "Failed to apply ArgoCD Application file $line to Kubernetes cluster!"
      exit 24
    fi
    app=`yq -r '.metadata.name' $line`
    argocd --server ${ARGOCD_HOSTNAME} --grpc-web --auth-token=${ARGO_TOKEN} app get --refresh $app
    argocd --server ${ARGOCD_HOSTNAME} --grpc-web --auth-token=${ARGO_TOKEN} app diff --exit-code=false $app
  done
  if [ $? -ne 0 ]
  then
    echo "Plan ArgoCD failed!"
    exit 20
  fi
elif [ "$COMMAND" == "argocd-apply" ]
then
  gcloud container clusters get-credentials $KUBERNETES_CLUSTER_NAME --region $GOOGLE_REGION --project $GOOGLE_PROJECT
  export ARGOCD_HOSTNAME=`kubectl get ingress -n argocd -l app.kubernetes.io/component=server  -o jsonpath='{.items[*].spec.rules[*].host}'`
  echo "ArgoCD hostname is $ARGOCD_HOSTNAME"
  export ARGO_TOKEN=`kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-infralib-token -o jsonpath="{.data.token}" | base64 -d`
  
  find . -type f -name '*.yaml' | while read line
  do
    app=`yq -r '.metadata.name' $line`
    argocd --server ${ARGOCD_HOST} --auth-token=${ARGOCD_TOKEN} app sync $app
    argocd --server ${ARGOCD_HOST} --auth-token=${ARGOCD_TOKEN} app wait --timeout 300 --health --sync --operation $app
  done
  if [ $? -ne 0 ]
  then
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
fi 

 
