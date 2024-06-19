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
  gcloud auth activate-service-account --key-file=application_default_credentials.json
  gcloud config set account $(gcloud auth list --filter=status:ACTIVE --format="value(account)")
  gcloud container clusters get-credentials $GKS_CLUSTER --region $GOOGLE_REGION --project $GOOGLE_PROJECT


  find . -type f -name '*.yaml' | while read line
  do
    kubectl apply -n argocd -f $line
    if [ $? -ne 0 ]
    then
      echo "Failed to apply ArgoCD Application file $line to Kubernetes cluster!"
      exit 24
    fi
    app=`echo $line | cut -d"." -f1`
    argocd --server ${ARGOCD_HOST} --auth-token=${ARGOCD_TOKEN} app get --refresh $app
    argocd --server ${ARGOCD_HOST} --auth-token=${ARGOCD_TOKEN} app diff --exit-code=false $app
  done
  if [ $? -ne 0 ]
  then
    echo "Plan ArgoCD failed!"
    exit 20
  fi
elif [ "$COMMAND" == "argocd-apply" ]
then
  find . -type f -name '*.yaml' | while read line
  do
    app=`echo $line | cut -d"." -f1`
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

 
