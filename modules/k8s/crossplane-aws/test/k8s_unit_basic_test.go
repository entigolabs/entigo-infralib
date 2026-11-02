package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/random"
)

// Crossplane runs provider deployments in its own install namespace (the
// crossplane-core application), not in this app's namespace.
const crossplaneNamespace = "crossplane-system"

func TestCrossplaneAws(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)
		releaseName := c.Namespace

		awsRegion := e.MustRegion(t)

		_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, c, releaseName, 60, 1*time.Second)
		require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

		// Install AWS provider
		provider, err := k8s.WaitUntilProviderAvailable(t, c, "upbound-provider-aws-s3", 60, 1*time.Second)
		require.NoError(t, err, "Provider aws error")
		assert.NotNil(t, provider, "Provider aws is nil")
		providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
		assert.NotEmpty(t, providerDeployment, "Provider aws currentRevision is empty")
		k8s.WaitUntilDeploymentAvailable(t, c.WithNamespace(crossplaneNamespace), providerDeployment, 60, 1*time.Second)

		err = k8s.WaitUntilResourcesAvailable(t, c, "aws.upbound.io/v1beta1", []string{"providerconfigs"}, 60, 1*time.Second)
		require.NoError(t, err, "Providerconfigs crd error")
		resource := schema.GroupVersionResource{Group: "aws.upbound.io", Version: "v1beta1", Resource: "providerconfigs"}
		_, err = k8s.WaitUntilProviderConfigAvailable(t, c, resource, releaseName, 60, 1*time.Second)
		require.NoError(t, err, "Provider config error")

		// Create S3 Bucket
		bucketName := fmt.Sprintf("entigo-infralib-test-%s-crossplane-%s", strings.ToLower(random.UniqueId()), e.Prefix)
		bucket, err := k8s.CreateK8SBucket(t, c, bucketName, "./templates/s3bucket.yaml")
		require.NoError(t, err, "Creating bucket error")
		assert.NotNil(t, bucket, "Bucket is nil")
		assert.Equal(t, bucketName, bucket.GetName(), "Bucket name is not equal")

		_, err = k8s.WaitUntilK8SBucketAvailable(t, c, bucketName, 30, 4*time.Second)
		if err != nil {
			_ = k8s.DeleteK8SBucket(t, c, bucketName)
		}
		require.NoError(t, err, "Bucket syncing error")
		err = aws.WaitUntilBucketExists(t, awsRegion, bucketName, 30, 4*time.Second)
		if err != nil {
			_ = k8s.DeleteK8SBucket(t, c, bucketName)
		}
		require.NoError(t, err, "S3 bucket creation error")

		err = k8s.DeleteK8SBucket(t, c, bucketName)
		require.NoError(t, err, "Deleting bucket error")

		err = aws.WaitUntilBucketDeleted(t, awsRegion, bucketName, 6, 10*time.Second)
		require.NoError(t, err, "S3 Bucket deletion error")
		err = k8s.WaitUntilK8SBucketDeleted(t, c, bucketName, 12, 5*time.Second)
		require.NoError(t, err, "Bucket didn't get deleted")
	})
}
