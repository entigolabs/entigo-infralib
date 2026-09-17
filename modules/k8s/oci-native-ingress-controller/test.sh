#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
# The controller Deployment comes from the vendored upstream subchart, which ships no
# NetworkPolicy - same as the crossplane and karpenter modules.
export KUBESCORE_EXTRA_OPTS="--ignore-test pod-networkpolicy"
exec $SCRIPTPATH/../../../common/test.sh "$@"
