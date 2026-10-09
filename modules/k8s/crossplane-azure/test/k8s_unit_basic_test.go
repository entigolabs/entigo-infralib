package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
	kubernetesErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestK8sCrossplaneAzureBiz(t *testing.T) {
	testK8sCrossplaneAzure(t, "azure", "biz")
}

func TestK8sCrossplaneAzurePri(t *testing.T) {
	testK8sCrossplaneAzure(t, "azure", "pri")
}

func testK8sCrossplaneAzure(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	releaseName := namespaceName

	_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, kubectlOptions, releaseName, 60, 1*time.Second)
	require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

	for _, name := range []string{"upbound-provider-family-azure", "provider-azure-managedidentity", "provider-azure-authorization", "provider-azure-storage"} {
		_, err = k8s.WaitUntilProviderAvailable(t, kubectlOptions, name, 100, 6*time.Second)
		require.NoError(t, err, "Provider %s error", name)
	}

	_, err = k8s.WaitUntilProviderConfigAvailable(t, kubectlOptions, schema.GroupVersionResource{Group: "azure.m.upbound.io", Version: "v1beta1", Resource: "clusterproviderconfigs"}, releaseName, 60, 6*time.Second)
	require.NoError(t, err, "ClusterProviderConfig error")

	// Create a WorkloadIdentity (Azure managed identity + role assignment)
	client, err := k8s.GetDynamicKubernetesClientFromOptionsE(t, kubectlOptions)
	require.NoError(t, err, "Kubernetes client error")
	workloadIdentityGVR := schema.GroupVersionResource{Group: "azure.entigo.com", Version: "v1alpha1", Resource: "workloadidentities"}
	name := fmt.Sprintf("wi-test-%s", strings.ToLower(random.UniqueId()))
	workloadIdentity := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "azure.entigo.com/v1alpha1",
		"kind":       "WorkloadIdentity",
		"metadata":   map[string]interface{}{"name": name, "namespace": namespaceName},
		"spec": map[string]interface{}{
			"serviceAccountName": name,
			"roleAssignments": []interface{}{
				map[string]interface{}{
					"role":  "Monitoring Reader",
					"scope": fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", azure.SubscriptionID(), azure.ResourceGroup(envName)),
				},
			},
		},
	}}
	_, err = client.Resource(workloadIdentityGVR).Namespace(namespaceName).Create(context.Background(), workloadIdentity, metaV1.CreateOptions{})
	require.NoError(t, err, "Creating WorkloadIdentity error")

	_, err = k8s.WaitUntilNamespacedCrossplaneResourceAvailable(t, kubectlOptions, workloadIdentityGVR, name, 90, 10*time.Second)
	if err != nil {
		_ = client.Resource(workloadIdentityGVR).Namespace(namespaceName).Delete(context.Background(), name, metaV1.DeleteOptions{})
	}
	require.NoError(t, err, "WorkloadIdentity syncing error")

	err = client.Resource(workloadIdentityGVR).Namespace(namespaceName).Delete(context.Background(), name, metaV1.DeleteOptions{})
	require.NoError(t, err, "Deleting WorkloadIdentity error")

	_, err = retry.DoWithRetryE(t, fmt.Sprintf("Wait for WorkloadIdentity %s to be deleted", name), 60, 10*time.Second, func() (string, error) {
		_, err := client.Resource(workloadIdentityGVR).Namespace(namespaceName).Get(context.Background(), name, metaV1.GetOptions{})
		if kubernetesErrors.IsNotFound(err) {
			return "deleted", nil
		}
		return "", fmt.Errorf("WorkloadIdentity %s still exists", name)
	})
	require.NoError(t, err, "WorkloadIdentity didn't get deleted")
}
