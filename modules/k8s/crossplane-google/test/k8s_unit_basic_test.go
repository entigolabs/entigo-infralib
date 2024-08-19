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
	commonGoogle "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/helm"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCrossplaneGoogleBiz(t *testing.T) {
	testK8sCrossplaneGoogle(t, "gke_entigo-infralib2_europe-north1_runner-main-biz", "biz")
}

func TestK8sCrossplaneGooglePri(t *testing.T) {
	testK8sCrossplaneGoogle(t, "gke_entigo-infralib2_europe-north1_runner-main-pri", "pri")
}

func testK8sCrossplaneGoogle(t *testing.T, contextName, envName string) {
	t.Parallel()
	spew.Dump("")

	helmChartPath, err := filepath.Abs("..")
	require.NoError(t, err)

	projectID := strings.ToLower(os.Getenv("GOOGLE_PROJECT"))
	prefix := strings.ToLower(os.Getenv("TF_VAR_prefix"))
	namespaceName := "crossplane-system"
	releaseName := "crossplane-google"

	extraArgs := make(map[string][]string)
	setValues := make(map[string]string)

	googleServiceAccount := commonGoogle.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-%s-service_account_email/versions/latest", projectID, envName))

	setValues["deploymentRuntimeConfig.googleServiceAccount"] = googleServiceAccount
	setValues["installProviderConfig"] = "false"
	setValues["google.projectID"] = projectID

	if prefix != "runner-main" {
		extraArgs["upgrade"] = []string{"--skip-crds"}
		extraArgs["install"] = []string{"--skip-crds"}
	}

	kubectlOptions := terrak8s.NewKubectlOptions(contextName, "", namespaceName)

	helmOptions := &helm.Options{
		SetValues:         setValues,
		KubectlOptions:    kubectlOptions,
		BuildDependencies: false,
		ExtraArgs:         extraArgs,
	}

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer helm.Delete(t, helmOptions, releaseName, true)
	}

	// Install DeploymentRuntimeConfig and Provider
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
	_, err = k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "upbound-provider-family-gcp", 60, 6*time.Second)
	require.NoError(t, err, "upbound-provider-family-gcp error")
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "provider-gcp-cloudplatform", 60, 6*time.Second)
	require.NoError(t, err, "provider-gcp-cloudplatform")
	_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, "provider-gcp-storage", 60, 6*time.Second)
	require.NoError(t, err, "provider-gcp-storage crd error")

	// Install ProviderConfig
	setValues["installProviderConfig"] = "true"
	helmOptions.SetValues = setValues
	helm.Upgrade(t, helmOptions, helmChartPath, releaseName)
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
