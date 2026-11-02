package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestLoki(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-read", 20, 6*time.Second), "loki-read deployment error")
		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace+"-gateway", 20, 6*time.Second), "loki-gateway deployment error")
		require.NoError(t, k8s.WaitUntilPodAvailableE(t, c, c.Namespace+"-write-0", 20, 6*time.Second), "loki-write-0 pod error")
		require.NoError(t, k8s.WaitUntilPodAvailableE(t, c, c.Namespace+"-backend-0", 20, 6*time.Second), "loki-backend-0 pod error")

		// Loki writes a cluster seed file to its store on startup, which proves the
		// pods can write to the bucket with their cloud identity.
		switch e.Cloud {
		case env.CloudAWS:
			bucket := fmt.Sprintf("%s-%s-%s-%s", e.Prefix, c.Namespace, aws.AccountID(t, e), e.MustRegion(t))
			require.NoError(t, aws.WaitUntilBucketFileAvailable(t, e.MustRegion(t), bucket, "loki_cluster_seed.json", 20, 6*time.Second), "File not found in AWS bucket")
		case env.CloudGoogle:
			require.NoError(t, google.WaitUntilBucketFileAvailable(t, fmt.Sprintf("%s-%s-logs", e.Prefix, c.Namespace), "loki_cluster_seed.json", 20, 6*time.Second), "File not found in Google bucket")
		case env.CloudAzure:
			account, err := k8s.RunKubectlAndGetOutputE(t, c, "get", "accounts.storage.azure.m.upbound.io", c.Namespace, "-o", `jsonpath={.metadata.annotations.crossplane\.io/external-name}`)
			require.NoError(t, err, "storage Account error")
			azure.WaitUntilBlobExists(t, azure.AccountNamed(account), fmt.Sprintf("%s-%s-logs", e.Prefix, c.Namespace), "loki_cluster_seed.json", 20, 6*time.Second)
		}

		k8s.WaitUntilRouteReachable(t, c, "", routeRetries(e), 6*time.Second)
	})
}

// Google and Azure gateways take long to program a new route.
func routeRetries(e *env.Environment) int {
	if e.Cloud == env.CloudGoogle || e.Cloud == env.CloudAzure {
		return 400
	}
	return 100
}
