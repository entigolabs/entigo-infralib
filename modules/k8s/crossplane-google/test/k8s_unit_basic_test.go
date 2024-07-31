package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCrossplaneBiz(t *testing.T) {
	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "runner-main-biz")
}

func TestK8sCrossplanePri(t *testing.T) {
	testK8sCrossplane(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "runner-main-pri")
}

func testK8sCrossplane(t *testing.T, contextName string, runnerName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	googleProjectID := strings.ToLower(os.Getenv("GOOGLE_PROJECT"))
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := "crossplane-system"
	releaseName := "crossplane-system"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	if prefix != "runner-main" {
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	setValues["installDeploymentRuntimeConfig"] = "false"
	setValues["installProvider"] = "false"
	setValues["installProviderConfig"] = "false"
	setValues["google.projectID"] = googleProjectID
	helmOptions := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
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

	googleServiceAccountID := fmt.Sprintf("%s-cp", runnerName)
	if len(runnerName) > 25 {
		googleServiceAccountID = fmt.Sprintf("%s-cp", runnerName[:26])
	}
	setValues["installDeploymentRuntimeConfig"] = "true"
	setValues["deploymentRuntimeConfig.googleServiceAccount"] = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", googleServiceAccountID, googleProjectID)
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, fmt.Sprintf("google-%s", releaseName), 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	setValues["installProvider"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "provider-gcp-storage", 60, 5*time.Second)
	require.NoError(t, err, "Providers crd error")
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "upbound-provider-family-gcp", 60, 5*time.Second)
	require.NoError(t, err, "Providers crd error")
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "provider-gcp-cloudplatform", 60, 5*time.Second)
	require.NoError(t, err, "Providers crd error")

	setValues["installProviderConfig"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, schema.GroupVersionResource{Group: "gcp.upbound.io", Version: "v1beta1", Resource: "providerconfigs"}, fmt.Sprintf("google-%s", releaseName), 60, 5*time.Second)
	require.NoError(t, err, "ProviderConfig crd error")

	// Create S3 bucket
	bucketName := "entigo-infralib-test" + "-" + strings.ToLower(random.UniqueId()) + "-" + releaseName
	bucket, err := k8s.CreateK8SBucket(t, kubectlOptions, bucketName, "./templates/bucket.yaml")
	require.NoError(t, err, "Creating bucket error")
	assert.NotNil(t, bucket, "Bucket is nil")
	assert.Equal(t, bucketName, bucket.GetName(), "Bucket name is not equal")

	_, err = k8s.WaitUntilK8SBucketAvailable(t, kubectlOptions, bucketName, 30, 4*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName) // Try to delete bucket
	}
	require.NoError(t, err, "Bucket syncing error")

	err = google.WaitUntilGCPBucketExists(t, bucketName, 30, 4*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName) // Try to delete bucket
	}
	require.NoError(t, err, "S3 bucket creation error")

	err = k8s.DeleteK8SBucket(t, kubectlOptions, bucketName)
	require.NoError(t, err, "Deleting bucket error")

	err = google.WaitUntilGCPBucketDeleted(t, bucketName, 6, 10*time.Second)
	require.NoError(t, err, "S3 Bucket deletion error")

	err = k8s.WaitUntilK8SBucketDeleted(t, kubectlOptions, bucketName, 12, 5*time.Second)
	require.NoError(t, err, "Bucket didn't get deleted")
}
