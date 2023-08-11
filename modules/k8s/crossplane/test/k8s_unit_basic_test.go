package test

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTerraformBasicBiz(t *testing.T) {
	testTerraformBasic(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz")
}

func TestTerraformBasicPri(t *testing.T) {
	testTerraformBasic(t, "arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-pri")
}

func testTerraformBasic(t *testing.T, contextName string) {
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := fmt.Sprintf("crossplane-system")
	releaseName := "crossplane"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		//releaseName = fmt.Sprintf("crossplane-%s", prefix)
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	kubectlOptions := k8s.NewKubectlOptions(contextName, "", namespaceName)

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
		//k8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	}

	err = k8s.CreateNamespaceE(t, kubectlOptions, namespaceName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			fmt.Println("Namespace already exists.")
		} else {
			t.Fatal("Error:", err)
		}
	}

	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "crossplane", 60, 1*time.Second)
	require.NoError(t, err, "Crossplane deployment error")
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "crossplane-rbac-manager", 60, 1*time.Second)
	require.NoError(t, err, "Crossplane-rbac-manager deployment error")
	err = WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1", []string{"providers"}, 60, 1*time.Second)
	require.NoError(t, err, "Providers crd error")
	err = WaitUntilResourcesAvailable(t, kubectlOptions, "pkg.crossplane.io/v1alpha1", []string{"controllerconfigs"}, 60, 1*time.Second)
	require.NoError(t, err, "Controllerconfigs crd error")

	setValues["installProvider"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	provider, err := WaitUntilProviderAvailable(t, kubectlOptions, "aws-crossplane", 60, 1*time.Second)
	require.NoError(t, err, "Provider error")
	assert.NotNil(t, provider, "Provider is nil")
	providerDeployment := GetStringValue(provider.Object, "status", "currentRevision")
	assert.NotEmpty(t, providerDeployment, "Provider currentRevision is empty")
	err = k8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, providerDeployment, 60, 1*time.Second)
	require.NoError(t, err, "Provider deployment error")
	_, err = WaitUntilControllerConfigAvailable(t, kubectlOptions, "aws-crossplane", 60, 1*time.Second)
	require.NoError(t, err, "Controller config error")

	setValues["installProviderConfig"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)

	err = WaitUntilResourcesAvailable(t, kubectlOptions, "aws.crossplane.io/v1beta1", []string{"providerconfigs"}, 60, 1*time.Second)
	require.NoError(t, err, "Providerconfigs crd error")
	_, err = WaitUntilProviderConfigAvailable(t, kubectlOptions, "aws-crossplane", 60, 1*time.Second)
	require.NoError(t, err, "Provider config error")

	bucketName := "entigo-infralib-test" + "-" + strings.ToLower(random.UniqueId()) + "-" + releaseName
	bucket, err := CreateS3Bucket(t, kubectlOptions, bucketName, "./templates/s3bucket.yaml")
	require.NoError(t, err, "Creating bucket error")
	assert.NotNil(t, bucket, "Bucket is nil")
	assert.Equal(t, bucketName, bucket.GetName(), "Bucket name is not equal")

	_, err = WaitUntilBucketAvailable(t, kubectlOptions, bucketName, 30, 2*time.Second)
	if err != nil {
		_ = DeleteBucket(t, kubectlOptions, bucketName) // Try to delete bucket
	}
	require.NoError(t, err, "Bucket syncing error")
	err = WaitUntilS3BucketExists(t, "eu-north-1", bucketName, 30, 2*time.Second)
	if err != nil {
		_ = DeleteBucket(t, kubectlOptions, bucketName) // Try to delete bucket
	}
	require.NoError(t, err, "S3 bucket creation error")

	err = DeleteBucket(t, kubectlOptions, bucketName)
	require.NoError(t, err, "Deleting bucket error")

	err = WaitUntilS3BucketDeleted(t, "eu-north-1", bucketName, 30, 2*time.Second)
	require.NoError(t, err, "S3 Bucket deletion error")
	err = WaitUntilBucketDeleted(t, kubectlOptions, bucketName, 30, 2*time.Second)
	require.NoError(t, err, "Bucket didn't get deleted")
}
