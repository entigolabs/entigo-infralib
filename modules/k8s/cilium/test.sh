#!/bin/bash
SCRIPTPATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $SCRIPTPATH
# The agent loads eBPF programs into the node's kernel, so it runs privileged. No resource limits:
# an agent killed for memory takes the node's pod networking down with it.
export KUBESCORE_EXTRA_OPTS="--ignore-test container-security-context-privileged --ignore-test container-resources --ignore-test container-ephemeral-storage-request-and-limit --ignore-test pod-probes --ignore-test pod-networkpolicy"
exec $SCRIPTPATH/../../../common/test.sh "$@"
