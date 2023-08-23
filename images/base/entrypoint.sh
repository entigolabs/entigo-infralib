#!/bin/bash
set -x

#temporary.
echo "I work"
exit 0

[ -z $PREFIX ] && echo "PREFIX must be set" && exit 1
[ -z $REGION ] && echo "REGION must be set" && exit 1
[ -z $PROJECT_NAME ] && echo "PROJECT_NAME must be set" && exit 1
[ -z $ENVIRONMENT ] && echo "ENVIRONMENT must be set" && exit 1
[ -z $PROJECT_GIT ] && echo "PROJECT_GIT must be set" && exit 1
[ -z $PROJECT_PATH ] && echo "PROJECT_PATH must be set" && exit 1
[ -z $PROJECT_TYPE ] && echo "PROJECT_TYPE must be set" && exit 1
[ -z $COMMAND ] && echo "COMMAND must be set" && exit 1
[ -z $ACCOUNT_ID ] && echo "ACCOUNT_ID must be set" && exit 1

export TF_VERSION="${TERRAFORM_VERSION:=1.0.11}"
export TF_IN_AUTOMATION=1
export GIT_SSH_COMMAND="ssh -i $(pwd)/sshkey -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"


echo "Downloading the terraform version ${TF_VERSION} binary from S3."
aws s3 cp s3://${PREFIX}-${PROJECT_NAME}-${ACCOUNT_ID}/terraform_${TF_VERSION} /bin/terraform --no-progress
if [ $? -ne 0 ]
then
  echo "Failed to fetch TF binary from S3. Downloading from public mirror!"
  wget -nv https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_linux_amd64.zip
  unzip terraform_${TF_VERSION}_linux_amd64.zip
  mv terraform /bin
  rm terraform_${TF_VERSION}_linux_amd64.zip
  echo "Copying terraform binary to S3 for future runs."
  aws s3 cp /bin/terraform s3://${PREFIX}-${PROJECT_NAME}-${ACCOUNT_ID}/terraform_${TF_VERSION} --no-progress --quiet
fi
chmod +x /bin/terraform




cd $CODEBUILD_SRC_DIR

if [ -f  sshkey.pub ]
then
  echo "For cloning using SSH key: $(cat sshkey.pub)"
fi
#PROJECT_TYPE
#shared - all envs use the same TF files
#branched - each env uses branch named after itself 
#pathed - each env uses folder named after itself 
if [ "$PROJECT_TYPE" == "shared" ]
then
  export BRANCH="main"
  export TFPATH="${PROJECT_PATH}"
elif [ "$PROJECT_TYPE" == "branched" ]
then
  export BRANCH=$ENVIRONMENT
  export TFPATH="${PROJECT_PATH}"
elif [ "$PROJECT_TYPE" == "pathed" ]
then
  export BRANCH="main"
  export TFPATH="${PROJECT_PATH}/${ENVIRONMENT}"
fi

export TF_VAR_prefix=${PREFIX}
export TF_VAR_region=${REGION}
export TF_VAR_project_name=${PROJECT_NAME}
export TF_VAR_project_network_name=${PROJECT_NETWORK_NAME}
export TF_VAR_project_account=${ACCOUNT_ID}


if [ "$COMMAND" == "plan" -o "$COMMAND" == "plan-destroy" ]
then
  if [ ! -d plan ]
  then
   git clone --single-branch --branch $BRANCH $PROJECT_GIT plan
    if [ ! -d plan ]
    then
      echo "Cloning failed from  $BRANCH $PROJECT_GIT"
      exit 3
    fi
    rm -rf plan/.git
  fi
  if [ ! -d "plan/$TFPATH" ]
  then
    echo "Unable to find path $TFPATH in git $PROJECT_GIT branch $BRANCH"
    exit 5
  fi
  cd plan/$TFPATH
  cp -r $CODEBUILD_SRC_DIR/modules/* ./
  #cp -r $CODEBUILD_SRC_DIR/.terraform ./
  if [ -f "$CODEBUILD_SRC_DIR/sshkey" ]
  then
    echo "Exposing sshkey to pipeline!"
    cp -p $CODEBUILD_SRC_DIR/sshkey ./
    cp -p $CODEBUILD_SRC_DIR/sshkey.pub ./
  fi

elif [ "$COMMAND" == "apply" -o "$COMMAND" == "apply-destroy" ]
then
  if [ ! -f $CODEBUILD_SRC_DIR_Plan/tf.tar.gz ]
  then
    echo "Unable to find artifacts from plan stage! $CODEBUILD_SRC_DIR_Plan/plan/tf.tar.gz"
    exit 4
  fi
  tar -xvzf $CODEBUILD_SRC_DIR_Plan/tf.tar.gz
  cd plan/$TFPATH
fi

if [ ! -f backend.tf ]
then
cat  <<EOF > backend.tf
terraform {
  backend "s3" {
    bucket = "${PREFIX}-${PROJECT_NAME}-${ACCOUNT_ID}"
    key    = "terraform.tfstate"
    dynamodb_table = "${PREFIX}-${PROJECT_NAME}-${ACCOUNT_ID}"
    encrypt = true
  }
}
EOF
fi

if [ ! -f provider.tf ]
then
cat  <<EOF > provider.tf

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~>4.9"
    }
  }
}

provider "aws" {
  region	= "$REGION"
  max_retries	= 10

EOF

if [ "$PROJECT_DEFAULT_TAGS" != "" ]
then
cat  <<EOF >> provider.tf

  default_tags {
    tags = {
      $PROJECT_DEFAULT_TAGS
    }
  }

EOF

fi


cat  <<EOF >> provider.tf
    ignore_tags {
      key_prefixes = ["kubernetes.io/cluster/"]
    }
}
EOF
fi


terraform init -input=false
if [ $? -ne 0 ]
then
  echo "Terraform init failed."
  exit 14
fi
terraform workspace select $ENVIRONMENT -no-color || terraform workspace new $ENVIRONMENT -no-color && terraform workspace select $ENVIRONMENT -no-color && terraform init -input=false || exit 2

if [ "$COMMAND" == "plan" ]
then
  terraform plan -no-color -out $ENVIRONMENT.tf-plan -input=false
  if [ $? -ne 0 ]
  then
    echo "Failed to create TF plan!"
    exit 6
  fi
  cd $CODEBUILD_SRC_DIR
  tar -czf tf.tar.gz plan
elif [ "$COMMAND" == "apply" ]
then
  terraform apply -no-color -input=false $ENVIRONMENT.tf-plan
  if [ $? -ne 0 ]
  then
    echo "Apply failed!"
    exit 11
  fi
elif [ "$COMMAND" == "plan-destroy" ]
then
  terraform plan -destroy -no-color -out $ENVIRONMENT.tf-plan -input=false
  if [ $? -ne 0 ]
  then
    echo "Failed to create TF destroy plan!"
    exit 6
  fi
  cd $CODEBUILD_SRC_DIR
  tar -czf tf.tar.gz plan
elif [ "$COMMAND" == "apply-destroy" ]
then
  terraform apply -no-color -input=false $ENVIRONMENT.tf-plan
  if [ $? -ne 0 ]
  then
    echo "Apply destroy failed!"
    exit 11
  fi
else
  echo "Unknown command: $COMMAND"
fi 

 
