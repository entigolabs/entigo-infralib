#!/bin/bash

if [ "$PR_BRANCH" != "" ]
then
prefix="`whoami`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`"
else
prefix="`whoami`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]' | cut -d"-" -f1-2`"
fi

DOCKER_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
  DOCKER_OPTS="-it"
fi

VALUES_OPTS=""
if [ -f "test/static_values.yaml" ]
then
  echo "Adding test/static_values.yaml to static tests."
  VALUES_OPTS="-f test/static_values.yaml"
fi

if [ "$KUBESCORE_EXTRA_OPTS" == "" ]
then
  KUBESCORE_EXTRA_OPTS=""
fi

docker run $DOCKER_OPTS --rm -v "$(pwd)":/project -w /project --entrypoint /bin/bash martivo/kube-score:latest -c "helm lint $VALUES_OPTS --strict ."
if [ $? -ne 0 ]
then
        echo "helm lint failed"
        exit 1
fi

docker run $DOCKER_OPTS --rm -v "$(pwd)":/project -w /project --entrypoint /bin/bash martivo/kube-score:latest -c "helm template $prefix $VALUES_OPTS --skip-tests --namespace $prefix . > /dev/null"
if [ $? -ne 0 ]
then
        echo "helm template failed"
        exit 2
fi


docker run $DOCKER_OPTS --rm -v "$(pwd)":/project -w /project --entrypoint /bin/bash martivo/kube-score:latest -c "helm template $prefix $VALUES_OPTS --skip-tests --namespace $prefix . | kube-score score --ignore-test container-image-pull-policy --ignore-test container-security-context-readonlyrootfilesystem --ignore-test deployment-has-poddisruptionbudget --ignore-test container-security-context-user-group-id --ignore-test statefulset-has-servicename $KUBESCORE_EXTRA_OPTS -"
if [ $? -ne 0 ]
then
	echo "kube-score failed"
        exit 3
fi
 
