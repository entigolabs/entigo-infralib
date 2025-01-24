package test

import (
	"fmt"
	"testing"
	"time"
	"strings"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCrossplaneGoogleBiz(t *testing.T) {
	testK8sCrossplaneGoogle(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz")
}

func TestK8sCrossplaneGooglePri(t *testing.T) {
	testK8sCrossplaneGoogle(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri")
}

func testK8sCrossplaneGoogle(t *testing.T, contextName, envName string) {
	t.Parallel()
	namespaceName := "crossplane-system"
	releaseName := "crossplane-google"
	kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)

	_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "upbound-provider-family-gcp", 60, 6*time.Second)
	require.NoError(t, err, "upbound-provider-family-gcp error")
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "provider-gcp-cloudplatform", 60, 6*time.Second)
	require.NoError(t, err, "provider-gcp-cloudplatform")
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "provider-gcp-storage", 60, 6*time.Second)
	require.NoError(t, err, "provider-gcp-storage crd error")

	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, schema.GroupVersionResource{Group: "gcp.upbound.io", Version: "v1beta1", Resource: "providerconfigs"}, releaseName, 60, 6*time.Second)
	require.NoError(t, err, "ProviderConfig crd error")

	// Create cloud storage bucket
	bucketName := fmt.Sprintf("entigo-infralib-test-%s-crossplane-runner-main-%s", strings.ToLower(random.UniqueId()), envName)
	bucket, err := k8s.CreateK8SBucket(t, kubectlOptions, bucketName, "./templates/bucket.yaml")
	require.NoError(t, err, "Creating bucket error")
	assert.NotNil(t, bucket, "Bucket is nil")
	assert.Equal(t, bucketName, bucket.GetName(), "Bucket name is not equal")

	_, err = k8s.WaitUntilK8SBucketAvailable(t, kubectlOptions, bucketName, 30, 6*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName)
	}
	require.NoError(t, err, "Bucket syncing error")

	err = google.WaitUntilGCPBucketExists(t, bucketName, 30, 6*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName)
	}
	require.NoError(t, err, "Cloud storage bucket creation error")

	err = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName)
	require.NoError(t, err, "Deleting bucket error")

	err = google.WaitUntilGCPBucketDeleted(t, bucketName, 10, 6*time.Second)
	require.NoError(t, err, "Cloud storage bucket deletion error")

	err = k8s.WaitUntilK8SBucketDeleted(t, kubectlOptions, bucketName, 12, 6*time.Second)
	require.NoError(t, err, "Bucket didn't get deleted")
}
