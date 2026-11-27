package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/random"
	"github.com/entigolabs/entigo-infralib-test/retry"
)

var (
	roleDefinitions = schema.GroupVersionResource{Group: "authz.entigo.com", Version: "v1alpha1", Resource: "roledefinitions"}
	roleGrants      = schema.GroupVersionResource{Group: "authz.entigo.com", Version: "v1alpha1", Resource: "rolegrants"}
)

func TestAuthzControllerCluster(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		err := k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace, 30, 10*time.Second)
		require.NoError(t, err, "%s deployment error", c.Namespace)

		// A grant is rendered only where Zones are served (platform-apis). The Google and
		// Azure test environments have none, so the deployment check is all there is.
		if e.Cloud == env.CloudGoogle || e.Cloud == env.CloudAzure {
			return
		}
		testRoleGrant(t, c)
	})
}

// testRoleGrant creates a RoleDefinition and a RoleGrant and checks what the controller
// writes for them: ClusterReady on the grant, the ClusterRole and ClusterRoleBinding, and
// the role's key in argocd-rbac-cm. Deleting the grant must remove all of it again.
func testRoleGrant(t *testing.T, c *k8s.Client) {
	ctx := context.Background()
	clientset, err := c.Clientset()
	require.NoError(t, err)
	dynamicClient, err := c.Dynamic()
	require.NoError(t, err)

	// The chart renders its config values into a ConfigMap named after the release.
	config, err := clientset.CoreV1().ConfigMaps(c.Namespace).Get(ctx, c.Namespace, metav1.GetOptions{})
	require.NoError(t, err, "controller config error")
	argocdNamespace := config.Data["ARGOCD_NAMESPACE"]
	require.NotEmpty(t, argocdNamespace, "ARGOCD_NAMESPACE is not set")

	// One name for the definition and the grant, unique per run: the module is installed
	// once per environment and pull request runs share it.
	name := fmt.Sprintf("infralib-test-%s", random.LowerId(6))
	entry := fmt.Sprintf("%s@%s/%s", name, config.Data["ORGANIZATION"], config.Data["WORKSPACE"])
	rulesName := fmt.Sprintf("entigo-rg-%s-cluster", name)
	policyKey := fmt.Sprintf("policy.%s.csv", name)

	definition := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "authz.entigo.com/v1alpha1",
		"kind":       "RoleDefinition",
		"metadata":   map[string]any{"name": name},
		"spec": map[string]any{"rules": []any{map[string]any{
			"apiGroups": []any{"argoproj.io"},
			"resources": []any{"applications"},
			"verbs":     []any{"get"},
			"scope":     "Cluster",
		}}},
	}}
	grant := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "authz.entigo.com/v1alpha1",
		"kind":       "RoleGrant",
		"metadata":   map[string]any{"name": name},
		"spec": map[string]any{
			"roleRef":  name,
			"subjects": []any{map[string]any{"kind": "Group", "name": "infralib-test"}},
		},
	}}

	t.Cleanup(func() {
		for _, resource := range []schema.GroupVersionResource{roleGrants, roleDefinitions} {
			err := dynamicClient.Resource(resource).Delete(ctx, name, metav1.DeleteOptions{})
			if err != nil && !apierrors.IsNotFound(err) {
				t.Errorf("deleting %s %s: %v", resource.Resource, name, err)
			}
		}
	})
	_, err = k8s.CreateObject(t, c, definition, "", roleDefinitions)
	require.NoError(t, err, "creating RoleDefinition error")
	_, err = k8s.CreateObject(t, c, grant, "", roleGrants)
	require.NoError(t, err, "creating RoleGrant error")

	_, err = k8s.WaitUntilObjectAvailable(t, c, roleGrants, "", name, isClusterReady, 30, 4*time.Second)
	require.NoError(t, err, "RoleGrant %s is not ClusterReady", name)

	role, err := clientset.RbacV1().ClusterRoles().Get(ctx, rulesName, metav1.GetOptions{})
	require.NoError(t, err, "ClusterRole error")
	granted := false
	for _, rule := range role.Rules {
		if len(rule.APIGroups) == 1 && rule.APIGroups[0] == "argoproj.io" &&
			len(rule.Resources) == 1 && rule.Resources[0] == "applications" &&
			len(rule.Verbs) == 1 && rule.Verbs[0] == "get" {
			granted = true
		}
	}
	require.True(t, granted, "ClusterRole %s rules %v lack the definition's rule", rulesName, role.Rules)

	binding, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, rulesName, metav1.GetOptions{})
	require.NoError(t, err, "ClusterRoleBinding error")
	require.Equal(t, rulesName, binding.RoleRef.Name)
	require.Len(t, binding.Subjects, 1)
	require.Equal(t, "Group", binding.Subjects[0].Kind)
	require.Equal(t, fmt.Sprintf("oidc:%s", entry), binding.Subjects[0].Name)

	policy := fmt.Sprintf("p, role:%s, applications, get, */*, allow\ng, %s, role:%s\n", name, entry, name)
	retry.DoWithRetry(t, fmt.Sprintf("%s in argocd-rbac-cm", policyKey), 30, 4*time.Second, func() (string, error) {
		rbac, err := clientset.CoreV1().ConfigMaps(argocdNamespace).Get(ctx, "argocd-rbac-cm", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		if rbac.Data[policyKey] != policy {
			return "", fmt.Errorf("%s is %q, want %q", policyKey, rbac.Data[policyKey], policy)
		}
		return "", nil
	})

	err = dynamicClient.Resource(roleGrants).Delete(ctx, name, metav1.DeleteOptions{})
	require.NoError(t, err, "deleting RoleGrant error")
	retry.DoWithRetry(t, fmt.Sprintf("objects of RoleGrant %s removed", name), 30, 4*time.Second, func() (string, error) {
		if _, err := clientset.RbacV1().ClusterRoles().Get(ctx, rulesName, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
			return "", fmt.Errorf("ClusterRole %s still exists (%v)", rulesName, err)
		}
		if _, err := clientset.RbacV1().ClusterRoleBindings().Get(ctx, rulesName, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
			return "", fmt.Errorf("ClusterRoleBinding %s still exists (%v)", rulesName, err)
		}
		rbac, err := clientset.CoreV1().ConfigMaps(argocdNamespace).Get(ctx, "argocd-rbac-cm", metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		if _, found := rbac.Data[policyKey]; found {
			return "", fmt.Errorf("%s is still in argocd-rbac-cm", policyKey)
		}
		return "", nil
	})
}

func isClusterReady(grant *unstructured.Unstructured) bool {
	conditions, _, _ := unstructured.NestedSlice(grant.Object, "status", "conditions")
	for _, item := range conditions {
		condition, ok := item.(map[string]any)
		if ok && condition["type"] == "ClusterReady" {
			return condition["status"] == "True"
		}
	}
	return false
}
