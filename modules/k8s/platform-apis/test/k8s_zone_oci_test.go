package test

import (
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

// testPlatformApisOCI creates a Zone with one OKE node pool and a namespace pinned to it, and
// checks that a Pod in the namespace runs on a node of that pool.
func testPlatformApisOCI(t *testing.T, e *env.Environment) {
	c := k8s.Connect(t, e)
	cluster, _ := setupKubectlClients(c.Context, e.Prefix)

	t.Run("packages", func(t *testing.T) {
		checkPlatformApisHaveRequiredPackages(t, cluster, ZoneConfigurationName, TenancyFunctionName)
	})
	if t.Failed() {
		return
	}

	t.Run("zone", func(t *testing.T) {
		applyFile(t, cluster, "./templates/oci_zone.yaml")
		waitSyncedAndReady(t, cluster, ZoneKind, OCIZoneName, 60, 10*time.Second)
		waitSyncedAndReadyByLabel(t, cluster, NodePoolKind, OCIZoneName, 60, 10*time.Second)
		waitZoneNodesReady(t, cluster, OCIZoneName)
	})
	if t.Failed() {
		return
	}

	t.Run("pod-on-pool", func(t *testing.T) {
		applyFile(t, cluster, "./templates/oci_zone_namespace.yaml")
		waitNamespacePoliciesReady(t, cluster, OCIZoneName, OCIZoneNamespace)
		if t.Failed() {
			return
		}
		applyFile(t, cluster, "./templates/oci_zone_pod.yaml")
		nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, OCIZoneNamespace)
		zonePool := OCIZoneName + "-" + OCIZonePool

		testPodsRunning(t, nsOpts, OCIZonePodName)
		testPodNodeSelector(t, nsOpts, OCIZonePodName, "tenancy.entigo.com/zone-pool", zonePool)

		pod, err := terrak8s.GetPodE(t, nsOpts, OCIZonePodName)
		require.NoError(t, err)
		nodePool, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "get", "node", pod.Spec.NodeName,
			"-o", `jsonpath={.metadata.labels.tenancy\.entigo\.com/zone-pool}`)
		require.NoError(t, err)
		require.Equal(t, zonePool, nodePool, "pod %s/%s runs on node %s outside pool %s",
			OCIZoneNamespace, OCIZonePodName, pod.Spec.NodeName, zonePool)
	})
}
