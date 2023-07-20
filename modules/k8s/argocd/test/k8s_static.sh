#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/.."
cd $SCRIPTPATH || exit 1

if [ "$PR_BRANCH" != "" ]
then
prefix="`whoami`-$PR_BRANCH"
else
prefix="`whoami`-`git rev-parse --abbrev-ref HEAD | tr '[:upper:]' '[:lower:]'`"
fi

#if [ -d test/template ]
#then
#	docker run -it --rm -w /test -v $(pwd)/test:/test ubuntu rm -rf /test/template
#fi
docker run -ti --rm -v $(pwd):/apps -w /apps alpine/helm:3.12.2  lint --strict .
if [ $? -ne 0 ]
then
        echo "helm lint failed"
        exit 2
fi
#docker run -ti --rm -v $(pwd):/apps -w /apps alpine/helm:3.12.2 template $prefix --skip-tests --output-dir test/template --namespace $prefix . > /dev/null
#if [ $? -ne 0 ]
#then
#        echo "helm template failed"
#        exit 4
#fi

#docker run -it --rm -v $(pwd)/test/template:/template stackrox/kube-linter:v0.6.4 lint /template --exclude no-read-only-root-fs --exclude dangling-service 
#if [ $? -ne 0 ]
#then
#	echo "kube-linter failed"
#	exit 1
#fi
docker run -it --rm -v "$(pwd)":/project -w /project --entrypoint /bin/bash martivo/kube-score:latest -c "helm template $prefix --skip-tests --namespace $prefix . | kube-score score --ignore-test container-image-pull-policy --ignore-test container-security-context-readonlyrootfilesystem --ignore-test deployment-has-poddisruptionbudget --ignore-test container-security-context-user-group-id --ignore-test statefulset-has-servicename -"
if [ $? -ne 0 ]
then
	echo "kube-score failed"
        exit 3
fi
