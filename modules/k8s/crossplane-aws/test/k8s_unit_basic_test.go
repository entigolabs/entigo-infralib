package test

import (
	"fmt"
	"testing"
	"time"
	"os"
	"strings"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terraaws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sCrossplaneAWSBiz(t *testing.T) {
	testK8sCrossplaneAWS(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sCrossplaneAWSPri(t *testing.T) {
	testK8sCrossplaneAWS(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}

func testK8sCrossplaneAWS(t *testing.T, contextName string, envName string) {
	t.Parallel()

	namespaceName := "crossplane-system"
	releaseName := "crossplane-aws"
	
	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)
	output, err := terrak8s.RunKubectlAndGetOutputE(t, kubectlOptions, "auth", "can-i", "get", "pods")
	require.NoError(t, err, "Unable to connect to context %s cluster %s", contextName, err)
	require.Equal(t, output, "yes")
	
	awsRegion := terraaws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	
	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	// Install AWS provider
	provider, err := k8s.WaitUntilProviderAvailable(t, kubectlOptions, "upbound-provider-aws-s3", 60, 1*time.Second)
	require.NoError(t, err, "Provider aws error")
	assert.NotNil(t, provider, "Provider aws is nil")
	providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
	assert.NotEmpty(t, providerDeployment, "Provider aws currentRevision is empty")
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, providerDeployment, 60, 1*time.Second)


	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "aws.crossplane.io/v1beta1", []string{"providerconfigs"}, 60, 1*time.Second)
	require.NoError(t, err, "Providerconfigs crd error")
	resource := schema.GroupVersionResource{Group: "aws.crossplane.io", Version: "v1beta1", Resource: "providerconfigs"}
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, resource, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "Provider config error")

	// Create S3 Bucket
	bucketName := fmt.Sprintf("entigo-infralib-test-%s-crossplane-%s", strings.ToLower(random.UniqueId()), envName)
	bucket, err := k8s.CreateK8SBucket(t, kubectlOptions, bucketName, "./templates/s3bucket.yaml")
	require.NoError(t, err, "Creating bucket error")
	assert.NotNil(t, bucket, "Bucket is nil")
	assert.Equal(t, bucketName, bucket.GetName(), "Bucket name is not equal")

	_, err = k8s.WaitUntilK8SBucketAvailable(t, kubectlOptions, bucketName, 30, 4*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName)
	}
	require.NoError(t, err, "Bucket syncing error")
	err = aws.WaitUntilAWSBucketExists(t, awsRegion, bucketName, 30, 4*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName)
	}
	require.NoError(t, err, "S3 bucket creation error")

	err = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName)
	require.NoError(t, err, "Deleting bucket error")

	err = aws.WaitUntilAWSBucketDeleted(t, awsRegion, bucketName, 6, 10*time.Second)
	require.NoError(t, err, "S3 Bucket deletion error")
	err = k8s.WaitUntilK8SBucketDeleted(t, kubectlOptions, bucketName, 12, 5*time.Second)
	require.NoError(t, err, "Bucket didn't get deleted")
}
