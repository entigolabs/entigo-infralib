#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/.."
cd $SCRIPTPATH || exit 1

if [ "$PR_BRANCH" != "" ]
then
prefix="`whoami`-`echo $PR_BRANCH | tr '[:upper:]' '[:lower:]'`"
else
prefix="`whoami`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]'`"
fi

DOCKER_OPTS=""
if [ "$GITHUB_ACTION" == "" ]
then
  DOCKER_OPTS="-it"
fi

docker run $DOCKER_OPTS --rm -v $(pwd):/apps -w /apps alpine/helm:3.12.2  lint --strict .
if [ $? -ne 0 ]
then
        echo "helm lint failed"
        exit 2
fi
docker run $DOCKER_OPTS --rm -v "$(pwd)":/project -w /project --entrypoint /bin/bash martivo/kube-score:latest -c "helm template $prefix --skip-tests --namespace $prefix . | kube-score score --ignore-test container-image-pull-policy --ignore-test container-security-context-readonlyrootfilesystem --ignore-test deployment-has-poddisruptionbudget --ignore-test container-security-context-user-group-id --ignore-test statefulset-has-servicename -"
if [ $? -ne 0 ]
then
	echo "kube-score failed"
        exit 3
fi
