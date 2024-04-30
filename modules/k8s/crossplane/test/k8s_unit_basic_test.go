package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terraaws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestK8sCrossplaneBiz(t *testing.T) {
	testK8sCrossplane(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz", "runner-main-biz")
}

func TestK8sCrossplanePri(t *testing.T) {
	testK8sCrossplane(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri", "runner-main-pri")
}

func testK8sCrossplane(t *testing.T, contextName string, runnerName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("crossplane-system")
	releaseName := "crossplane-system"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)
	
	awsRegion := terraaws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	iamrole := terraaws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s/iam_role",runnerName))
	setValues["awsRole"] = iamrole
	
	if prefix != "runner-main" {
		//releaseName = fmt.Sprintf("crossplane-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
		
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)


	setValues["installProvider"] = "false"
	setValues["installProviderConfig"] = "false"
	helmOptions := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
		//terrak8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	}

	err = terrak8s.CreateNamespaceE(t, kubectlOptions, namespaceName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Namespace already exists.")
		} else {
			t.Fatal("Error:", err)
		}
	}

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "crossplane", 10, 5*time.Second)
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, "crossplane-rbac-manager", 10, 5*time.Second)
	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1", []string{"providers"}, 60, 1*time.Second)
	require.NoError(t, err, "Providers crd error")
	
	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1beta1", []string{"deploymentruntimeconfigs"}, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfig crd error")

	setValues["installProvider"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	provider, err := k8s.WaitUntilProviderAvailable(t, kubectlOptions, fmt.Sprintf("aws-%s", releaseName), 60, 1*time.Second)
	require.NoError(t, err, "Provider error")
	assert.NotNil(t, provider, "Provider is nil")
	providerDeployment := k8s.GetStringValue(provider.Object, "status", "currentRevision")
	assert.NotEmpty(t, providerDeployment, "Provider currentRevision is empty")
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, providerDeployment, 60, 1*time.Second)
	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, fmt.Sprintf("aws-%s", releaseName), 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	setValues["installProviderConfig"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	err = k8s.WaitUntilResourcesAvailable(t, kubectlOptions, "aws.crossplane.io/v1beta1", []string{"providerconfigs"}, 60, 1*time.Second)
	require.NoError(t, err, "Providerconfigs crd error")
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, fmt.Sprintf("aws-%s", releaseName), 60, 1*time.Second)
	require.NoError(t, err, "Provider config error")

	bucketName := "entigo-infralib-test" + "-" + strings.ToLower(random.UniqueId()) + "-" + releaseName
	bucket, err := k8s.CreateK8SBucket(t, kubectlOptions, bucketName, "./templates/s3bucket.yaml")
	require.NoError(t, err, "Creating bucket error")
	assert.NotNil(t, bucket, "Bucket is nil")
	assert.Equal(t, bucketName, bucket.GetName(), "Bucket name is not equal")

	_, err = k8s.WaitUntilK8SBucketAvailable(t, kubectlOptions, bucketName, 30, 2*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName) // Try to delete bucket
	}
	require.NoError(t, err, "Bucket syncing error")
	err = aws.WaitUntilAWSBucketExists(t, awsRegion, bucketName, 30, 2*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName) // Try to delete bucket
	}
	require.NoError(t, err, "S3 bucket creation error")

	err = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName)
	require.NoError(t, err, "Deleting bucket error")

	err = aws.WaitUntilAWSBucketDeleted(t, awsRegion, bucketName, 6, 10*time.Second)
	require.NoError(t, err, "S3 Bucket deletion error")
	err = k8s.WaitUntilK8SBucketDeleted(t, kubectlOptions, bucketName, 12, 5*time.Second)
	require.NoError(t, err, "Bucket didn't get deleted")
}
