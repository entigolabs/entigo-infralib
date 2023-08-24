#!/bin/bash
set -x

[ -z $TF_VAR_prefix ] && echo "TF_VAR_prefix must be set" && exit 1
[ -z $AWS_REGION ] && echo "AWS_REGION must be set" && exit 1
[ -z $COMMAND ] && echo "COMMAND must be set" && exit 1
[ -z $WORKSPACE ] && echo "WORKSPACE must be set" && exit 1
export TF_IN_AUTOMATION=1
cd $CODEBUILD_SRC_DIR

if [ -d ".git" ]
then
rm -rf .git
fi

if [ "$COMMAND" == "plan" -o "$COMMAND" == "plan-destroy" ]
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
  if [ ! -f $CODEBUILD_SRC_DIR_Plan/tf.tar.gz ]
  then
    echo "Unable to find artifacts from plan stage! $CODEBUILD_SRC_DIR_Plan/tf.tar.gz"
    exit 4
  fi
  tar -xvzf $CODEBUILD_SRC_DIR_Plan/tf.tar.gz
  cd "$TF_VAR_prefix/$WORKSPACE"
fi


terraform init -input=false -backend-config=../backend.conf
if [ $? -ne 0 ]
then
  echo "Terraform init failed."
  exit 14
fi
terraform workspace select $WORKSPACE -no-color || terraform workspace new $WORKSPACE -no-color && terraform workspace select $WORKSPACE -no-color && terraform init -input=false || exit 2

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
elif [ "$COMMAND" == "apply-destroy" ]
then
  terraform apply -no-color -input=false $WORKSPACE.tf-plan
  if [ $? -ne 0 ]
  then
    echo "Apply destroy failed!"
    exit 11
  fi
else
  echo "Unknown command: $COMMAND"
fi 

 
