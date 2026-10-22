package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/random"
)

func TestCrossplaneAzure(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)
		releaseName := c.Namespace

		_, err := k8s.WaitUntilDeploymentRuntimeConfigAvailable(t, c, releaseName, 60, 1*time.Second)
		require.NoError(t, err, "DeploymentRuntimeConfigAvailable error")

		for _, name := range []string{"upbound-provider-family-azure", "provider-azure-managedidentity", "provider-azure-authorization", "provider-azure-storage"} {
			_, err = k8s.WaitUntilProviderAvailable(t, c, name, 100, 6*time.Second)
			require.NoError(t, err, "Provider %s error", name)
		}

		_, err = k8s.WaitUntilProviderConfigAvailable(t, c, schema.GroupVersionResource{Group: "azure.m.upbound.io", Version: "v1beta1", Resource: "clusterproviderconfigs"}, releaseName, 60, 6*time.Second)
		require.NoError(t, err, "ClusterProviderConfig error")

		resourceGroup := azure.ResourceGroup(t, e)

		// Provider only: a managed identity without the composition
		identityGVR := schema.GroupVersionResource{Group: "managedidentity.azure.m.upbound.io", Version: "v1beta1", Resource: "userassignedidentities"}
		identityName := fmt.Sprintf("mr-test-%s", strings.ToLower(random.UniqueId()))
		identity := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "managedidentity.azure.m.upbound.io/v1beta1",
			"kind":       "UserAssignedIdentity",
			"metadata":   map[string]any{"name": identityName, "namespace": c.Namespace},
			"spec": map[string]any{
				"forProvider": map[string]any{
					"name":              fmt.Sprintf("%s-%s", e.Prefix, identityName),
					"location":          e.MustRegion(t),
					"resourceGroupName": resourceGroup,
				},
				"providerConfigRef": map[string]any{"kind": "ClusterProviderConfig", "name": releaseName},
			},
		}}
		_, err = c.CreateObjectE(identityGVR, c.Namespace, identity)
		require.NoError(t, err, "Creating UserAssignedIdentity error")

		_, err = k8s.WaitUntilNamespacedCrossplaneResourceAvailable(t, c, identityGVR, identityName, 60, 6*time.Second)
		if err != nil {
			_ = c.DeleteObjectE(identityGVR, c.Namespace, identityName)
		}
		require.NoError(t, err, "UserAssignedIdentity syncing error")

		err = c.DeleteObjectE(identityGVR, c.Namespace, identityName)
		require.NoError(t, err, "Deleting UserAssignedIdentity error")

		err = k8s.WaitUntilObjectDeleted(t, c, identityGVR, c.Namespace, identityName, 30, 6*time.Second)
		require.NoError(t, err, "UserAssignedIdentity didn't get deleted")

		// Composition: a WorkloadIdentity (managed identity, federated credential, role assignment, service account)
		workloadIdentityGVR := schema.GroupVersionResource{Group: "iam.entigo.com", Version: "v1alpha1", Resource: "workloadidentities"}
		name := fmt.Sprintf("wi-test-%s", strings.ToLower(random.UniqueId()))
		workloadIdentity := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "iam.entigo.com/v1alpha1",
			"kind":       "WorkloadIdentity",
			"metadata":   map[string]any{"name": name, "namespace": c.Namespace},
			"spec": map[string]any{
				"serviceAccountName": name,
				"azurePermissions": []any{
					map[string]any{
						"role":  "Monitoring Reader",
						"scope": fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", e.MustSubscriptionID(t), resourceGroup),
					},
				},
			},
		}}
		_, err = c.CreateObjectE(workloadIdentityGVR, c.Namespace, workloadIdentity)
		require.NoError(t, err, "Creating WorkloadIdentity error")

		_, err = k8s.WaitUntilNamespacedCrossplaneResourceAvailable(t, c, workloadIdentityGVR, name, 90, 10*time.Second)
		if err != nil {
			_ = c.DeleteObjectE(workloadIdentityGVR, c.Namespace, name)
		}
		require.NoError(t, err, "WorkloadIdentity syncing error")

		err = c.DeleteObjectE(workloadIdentityGVR, c.Namespace, name)
		require.NoError(t, err, "Deleting WorkloadIdentity error")

		err = k8s.WaitUntilObjectDeleted(t, c, workloadIdentityGVR, c.Namespace, name, 60, 10*time.Second)
		require.NoError(t, err, "WorkloadIdentity didn't get deleted")
	})
}
