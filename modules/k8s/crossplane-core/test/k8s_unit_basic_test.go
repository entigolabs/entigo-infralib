package test

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

// Crossplane records an UpdatedExternalResource event on a managed resource each
// time a provider updates the external object. A managed resource that keeps
// updating is a drift loop: the provider reads back something the spec does not
// state and re-applies the spec on every reconcile, as the S3 encryption
// configurations did when the fields the API returns were missing from the
// spec. Such a loop floods the cloud API and its audit trail, so the test
// counts the updates of every managed resource over a window and fails on one
// that is still being updated. A resource under real change updates a few
// times; a loop updates several times a second.
const (
	updateLoopWindow    = 2 * time.Minute
	updateLoopMaxUpdate = 10 // updates of one managed resource within the window
)

func TestCrossplaneCore(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		k8s.WaitUntilDeploymentAvailable(t, c, "crossplane", 10, 6*time.Second)
		k8s.WaitUntilDeploymentAvailable(t, c, "crossplane-rbac-manager", 10, 6*time.Second)

		err := k8s.WaitUntilResourcesAvailable(t, c, "pkg.crossplane.io/v1", []string{"providers"}, 60, 6*time.Second)
		require.NoError(t, err, "Providers crd error")

		err = k8s.WaitUntilResourcesAvailable(t, c, "pkg.crossplane.io/v1beta1", []string{"deploymentruntimeconfigs"}, 60, 6*time.Second)
		require.NoError(t, err, "DeploymentRuntimeConfig crd error")

		testNoUpdateLoops(t, c)
	})
}

// testNoUpdateLoops fails when a managed resource was updated more than
// updateLoopMaxUpdate times during updateLoopWindow.
func testNoUpdateLoops(t *testing.T, c *k8s.Client) {
	before := updatedExternalResourceCounts(t, c)
	time.Sleep(updateLoopWindow)
	after := updatedExternalResourceCounts(t, c)

	var looping []string
	for object, count := range after {
		if updates := count - before[object]; updates > updateLoopMaxUpdate {
			looping = append(looping, fmt.Sprintf("%s: %d updates in %s", object, updates, updateLoopWindow))
		}
	}
	sort.Strings(looping)
	require.Empty(t, looping, "managed resources are updated continuously, the provider re-applies their spec on every reconcile")
	t.Logf("no managed resource updated more than %d times in %s (%d resources with updates)", updateLoopMaxUpdate, updateLoopWindow, len(after))
}

// updatedExternalResourceCounts sums the UpdatedExternalResource events of the
// cluster per managed resource. Kubernetes aggregates repeated events of an
// object into one event with a count, and starts a new event after a while, so
// the counts of all events of a resource are added up.
func updatedExternalResourceCounts(t *testing.T, c *k8s.Client) map[string]int32 {
	clientset, err := c.Clientset()
	require.NoError(t, err, "kubernetes client error")
	events, err := clientset.CoreV1().Events(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: "reason=UpdatedExternalResource",
	})
	require.NoError(t, err, "listing UpdatedExternalResource events")

	counts := map[string]int32{}
	for _, event := range events.Items {
		object := event.InvolvedObject
		key := object.Kind + "/" + object.Name
		if object.Namespace != "" {
			key = object.Kind + "/" + object.Namespace + "/" + object.Name
		}
		count := event.Count
		if event.Series != nil && event.Series.Count > count {
			count = event.Series.Count
		}
		if count == 0 {
			count = 1
		}
		counts[key] += count
	}
	return counts
}

