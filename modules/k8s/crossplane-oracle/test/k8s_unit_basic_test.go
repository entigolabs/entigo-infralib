package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/entigolabs/entigo-infralib-common/oracle"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TODO(oracle): re-enable once a shared Oracle OKE test cluster exists in CI.
// There is no OKE cluster behind this test yet, so it only ever fails in
// CheckKubectlConnection.
/*
func TestK8sCrossplaneOracleDev(t *testing.T) {
	testK8sCrossplaneOracle(t, "oracle", "dev")
}
*/

func testK8sCrossplaneOracle(t *testing.T, cloudName string, envName string) {
	t.Parallel()

	releaseName := "crossplane-oracle"

	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)
	// Crossplane runs provider deployments in its own install namespace, not in this
	// app's destination namespace (only the credentials secret lives there).
	kubectlOptions.Namespace = "crossplane-system"

	_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 2*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	// The family provider must be installed and healthy before per-service providers.
	provider, err := k8s.WaitUntilProviderAvailable(t, kubectlOptions, "oracle-provider-family-oci", 120, 2*time.Second)
	require.NoError(t, err, "Provider family oci error")
	assert.NotNil(t, provider, "Provider family oci is nil")
	providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
	assert.NotEmpty(t, providerDeployment, "Provider family oci currentRevision is empty")
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, providerDeployment, 60, 2*time.Second)

	// identity carries the per-app Policy MRs the other oracle modules depend on.
	identityProvider, err := k8s.WaitUntilProviderAvailable(t, kubectlOptions, "oracle-provider-oci-identity", 120, 2*time.Second)
	require.NoError(t, err, "Provider oci identity error")
	assert.NotNil(t, identityProvider, "Provider oci identity is nil")

	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "oci.upbound.io/v1beta1", []string{"providerconfigs"}, 60, 2*time.Second)
	require.NoError(t, err, "Providerconfigs crd error")
	resource := schema.GroupVersionResource{Group: "oci.upbound.io", Version: "v1beta1", Resource: "providerconfigs"}
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, resource, releaseName, 60, 2*time.Second)
	require.NoError(t, err, "Provider config error")

	// Everything above proves the provider installed. Provision a real bucket to prove it
	// can actually reach OCI - a provider can be Healthy and still fail on credentials or IAM.
	region := os.Getenv("OCI_REGION")
	require.NotEmpty(t, region, "OCI_REGION must be set")
	compartmentId := os.Getenv("ORACLE_COMPARTMENT_ID")
	require.NotEmpty(t, compartmentId, "ORACLE_COMPARTMENT_ID must be set")

	// Tenancy-wide and not derivable from the compartment, so it has to be looked up.
	objectStorageNamespace, err := oracle.GetObjectStorageNamespace(region)
	require.NoError(t, err, "Getting object storage namespace error")

	// Bucket names are unique per Object Storage namespace, so randomise to let parallel
	// runs coexist.
	bucketName := fmt.Sprintf("entigo-infralib-test-%s-crossplane", strings.ToLower(random.UniqueId()))
	bucketResource := schema.GroupVersionResource{Group: "objectstorage.oci.upbound.io", Version: "v1alpha1", Resource: "buckets"}

	bucketObject, err := k8s.ReadObjectFromFile(t, "./templates/bucket.yaml")
	require.NoError(t, err, "Reading bucket template error")
	bucketObject.SetName(bucketName)
	require.NoError(t, unstructured.SetNestedField(bucketObject.Object, bucketName, "spec", "forProvider", "name"))
	require.NoError(t, unstructured.SetNestedField(bucketObject.Object, objectStorageNamespace, "spec", "forProvider", "namespace"))
	require.NoError(t, unstructured.SetNestedField(bucketObject.Object, compartmentId, "spec", "forProvider", "compartmentId"))

	bucket, err := k8s.CreateObject(t, kubectlOptions, bucketObject, "", bucketResource)
	require.NoError(t, err, "Creating bucket error")
	assert.NotNil(t, bucket, "Bucket is nil")

	// Ready and Synced, so the bucket exists in OCI rather than just having been accepted.
	_, err = k8s.WaitUntilCrossplaneResourceAvailable(t, kubectlOptions, bucketResource, bucketName, 30, 4*time.Second)
	if err != nil {
		_ = k8s.DeleteCrossplaneResource(t, kubectlOptions, bucketResource, bucketName)
	}
	require.NoError(t, err, "Bucket syncing error")

	err = oracle.WaitUntilOCIBucketExists(t, region, bucketName, 30, 4*time.Second)
	if err != nil {
		_ = k8s.DeleteCrossplaneResource(t, kubectlOptions, bucketResource, bucketName)
	}
	require.NoError(t, err, "Bucket creation error")

	err = k8s.DeleteCrossplaneResource(t, kubectlOptions, bucketResource, bucketName)
	require.NoError(t, err, "Deleting bucket error")

	err = oracle.WaitUntilOCIBucketDeleted(t, region, bucketName, 30, 4*time.Second)
	require.NoError(t, err, "Bucket deletion error")
	err = k8s.WaitUntilCrossplaneResourceDeleted(t, kubectlOptions, bucketResource, bucketName, 12, 5*time.Second)
	require.NoError(t, err, "Bucket object didn't get deleted")
}
