package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/oracle"
	"github.com/entigolabs/entigo-infralib-test/random"
)

// Crossplane runs provider deployments in its own install namespace (the
// crossplane-core application), not in this app's destination namespace
// (only the credentials secret lives there).
const crossplaneNamespace = "crossplane-system"

func TestCrossplaneOracle(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)
		releaseName := c.Namespace

		_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, c, releaseName, 60, 2*time.Second)
		require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

		// The family provider must be installed and healthy before per-service providers.
		provider, err := k8s.WaitUntilProviderAvailable(t, c, "oracle-provider-family-oci", 120, 2*time.Second)
		require.NoError(t, err, "Provider family oci error")
		assert.NotNil(t, provider, "Provider family oci is nil")
		providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
		assert.NotEmpty(t, providerDeployment, "Provider family oci currentRevision is empty")
		k8s.WaitUntilDeploymentAvailable(t, c.WithNamespace(crossplaneNamespace), providerDeployment, 60, 2*time.Second)

		// identity carries the per-app Policy MRs the other oracle modules depend on.
		identityProvider, err := k8s.WaitUntilProviderAvailable(t, c, "oracle-provider-oci-identity", 120, 2*time.Second)
		require.NoError(t, err, "Provider oci identity error")
		assert.NotNil(t, identityProvider, "Provider oci identity is nil")

		err = k8s.WaitUntilResourcesAvailable(t, c, "oci.upbound.io/v1beta1", []string{"providerconfigs"}, 60, 2*time.Second)
		require.NoError(t, err, "Providerconfigs crd error")
		resource := schema.GroupVersionResource{Group: "oci.upbound.io", Version: "v1beta1", Resource: "providerconfigs"}
		_, err = k8s.WaitUntilProviderConfigAvailable(t, c, resource, releaseName, 60, 2*time.Second)
		require.NoError(t, err, "Provider config error")

		observers := []struct {
			resource schema.GroupVersionResource
			name     string
		}{
			{schema.GroupVersionResource{Group: "networking.oci.upbound.io", Version: "v1alpha1", Resource: "vcns"}, "vcn"},
			{schema.GroupVersionResource{Group: "networking.oci.upbound.io", Version: "v1alpha1", Resource: "subnets"}, "private-0"},
			{schema.GroupVersionResource{Group: "networking.oci.upbound.io", Version: "v1alpha1", Resource: "networksecuritygroups"}, "node"},
			{schema.GroupVersionResource{Group: "containerengine.oci.upbound.io", Version: "v1alpha1", Resource: "clusters"}, "oke"},
		}
		for _, step := range e.Steps {
			for _, m := range step.Modules {
				if m.Source == "oracle/kms" {
					observers = append(observers, struct {
						resource schema.GroupVersionResource
						name     string
					}{schema.GroupVersionResource{Group: "kms.oci.upbound.io", Version: "v1alpha1", Resource: "keys"}, "data"})
				}
			}
		}
		for _, o := range observers {
			_, err = k8s.WaitUntilCrossplaneResourceAvailable(t, c, o.resource, o.name, 90, 4*time.Second)
			require.NoError(t, err, fmt.Sprintf("Observer %s %s error", o.resource.Resource, o.name))
		}

		// Everything above proves the provider installed. Provision a real bucket to prove it
		// can actually reach OCI - a provider can be Healthy and still fail on credentials or IAM.
		region := e.MustRegion(t)
		compartmentId := e.MustCompartmentID(t)

		// Tenancy-wide and not derivable from the compartment, so it has to be looked up.
		objectStorageNamespace, err := oracle.ObjectStorageNamespace(region)
		require.NoError(t, err, "Getting object storage namespace error")

		// Bucket names are unique per Object Storage namespace, so randomise to let parallel
		// runs coexist.
		bucketName := fmt.Sprintf("entigo-infralib-test-%s-crossplane", strings.ToLower(random.UniqueId()))
		bucketResource := k8s.OracleBuckets

		bucketObject, err := k8s.ReadObjectFromFile(t, "./templates/bucket.yaml")
		require.NoError(t, err, "Reading bucket template error")
		bucketObject.SetName(bucketName)
		require.NoError(t, unstructured.SetNestedField(bucketObject.Object, bucketName, "spec", "forProvider", "name"))
		require.NoError(t, unstructured.SetNestedField(bucketObject.Object, objectStorageNamespace, "spec", "forProvider", "namespace"))
		require.NoError(t, unstructured.SetNestedField(bucketObject.Object, compartmentId, "spec", "forProvider", "compartmentId"))

		bucket, err := k8s.CreateObject(t, c, bucketObject, "", bucketResource)
		require.NoError(t, err, "Creating bucket error")
		assert.NotNil(t, bucket, "Bucket is nil")

		// Ready and Synced, so the bucket exists in OCI rather than just having been accepted.
		_, err = k8s.WaitUntilCrossplaneResourceAvailable(t, c, bucketResource, bucketName, 90, 4*time.Second)
		if err != nil {
			_ = k8s.DeleteCrossplaneResource(t, c, bucketResource, bucketName)
		}
		require.NoError(t, err, "Bucket syncing error")

		err = oracle.WaitUntilBucketExists(t, region, bucketName, 30, 4*time.Second)
		if err != nil {
			_ = k8s.DeleteCrossplaneResource(t, c, bucketResource, bucketName)
		}
		require.NoError(t, err, "Bucket creation error")

		err = k8s.DeleteCrossplaneResource(t, c, bucketResource, bucketName)
		require.NoError(t, err, "Deleting bucket error")

		err = oracle.WaitUntilBucketDeleted(t, region, bucketName, 90, 4*time.Second)
		require.NoError(t, err, "Bucket deletion error")
		err = k8s.WaitUntilCrossplaneResourceDeleted(t, c, bucketResource, bucketName, 36, 5*time.Second)
		require.NoError(t, err, "Bucket object didn't get deleted")
	})
}
