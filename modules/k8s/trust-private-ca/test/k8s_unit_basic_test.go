package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	kubernetesErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

var (
	clusterPolicyGVR = schema.GroupVersionResource{Group: "kyverno.io", Version: "v1", Resource: "clusterpolicies"}
	configMapGVR     = schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
	namespaceGVR     = schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}
	podGVR           = schema.GroupVersionResource{Version: "v1", Resource: "pods"}
)

func TestTrustPrivateCa(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"aws_biz":    testExistingNamespace,
		"aws_pri":    testNewNamespace,
		"azure_biz":  testExistingNamespace,
		"azure_pri":  testNewNamespace,
		"google_biz": testExistingNamespace,
		"google_pri": testNewNamespace,
		"oracle_biz": testExistingNamespace,
		"oracle_pri": testNewNamespace,
	})
}

// biz lists the default namespace, which exists before the module is installed.
func testExistingNamespace(t *testing.T, e *env.Environment) {
	c := k8s.Connect(t, e)
	waitUntilPoliciesReady(t, c)

	configMap, err := k8s.WaitUntilObjectAvailable(t, c, configMapGVR, "default", c.Namespace, nil, 30, 10*time.Second)
	require.NoError(t, err, "ConfigMap was not copied to the default namespace")
	require.Equal(t, 1, certificates(configMap, c.Namespace), "ConfigMap must hold the one certificate of test/<env>.yaml")
}

// pri lists a namespace that the test creates, and checks a pod created there.
func testNewNamespace(t *testing.T, e *env.Environment) {
	c := k8s.Connect(t, e)
	waitUntilPoliciesReady(t, c)

	namespace := fmt.Sprintf("%s-test", c.Namespace)
	_, err := c.CreateObjectE(namespaceGVR, "", &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata":   map[string]any{"name": namespace},
	}})
	if !kubernetesErrors.IsAlreadyExists(err) {
		require.NoError(t, err, "Creating namespace %s error", namespace)
	}
	t.Cleanup(func() {
		require.NoError(t, c.DeleteObjectE(namespaceGVR, "", namespace), "Deleting namespace %s error", namespace)
	})

	configMap, err := k8s.WaitUntilObjectAvailable(t, c, configMapGVR, namespace, c.Namespace, nil, 30, 10*time.Second)
	require.NoError(t, err, "ConfigMap was not copied to the new namespace %s", namespace)
	require.Equal(t, 2, certificates(configMap, c.Namespace), "ConfigMap must hold both certificates of test/<env>.yaml")

	// The policy changes the pod when it is created; the pod does not have to start
	pod, err := c.CreateObjectE(podGVR, namespace, &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata":   map[string]any{"name": "trust"},
		"spec": map[string]any{
			"containers": []any{map[string]any{"name": "trust", "image": "registry.k8s.io/pause:3.10"}},
		},
	}})
	require.NoError(t, err, "Creating a pod in %s error", namespace)
	containers, _, _ := unstructured.NestedSlice(pod.Object, "spec", "containers")
	require.Len(t, containers, 1)
	mounts, _, _ := unstructured.NestedSlice(containers[0].(map[string]any), "volumeMounts")
	mountPath := fmt.Sprintf("/usr/local/share/ca-certificates/%s.crt", c.Namespace)
	mounted := false
	for _, mount := range mounts {
		if mount.(map[string]any)["mountPath"] == mountPath {
			mounted = true
		}
	}
	require.True(t, mounted, "pod created in %s has no mount %s", namespace, mountPath)
}

func waitUntilPoliciesReady(t *testing.T, c *k8s.Client) {
	for _, policy := range []string{"copy", "mount"} {
		_, err := k8s.WaitUntilObjectAvailable(t, c, clusterPolicyGVR, "", fmt.Sprintf("%s-%s", c.Namespace, policy), isReady, 30, 10*time.Second)
		require.NoError(t, err, "ClusterPolicy %s error", policy)
	}
}

func isReady(object *unstructured.Unstructured) bool {
	conditions, _, _ := unstructured.NestedSlice(object.Object, "status", "conditions")
	for _, item := range conditions {
		condition, ok := item.(map[string]any)
		if ok && condition["type"] == "Ready" && condition["status"] == "True" {
			return true
		}
	}
	return false
}

func certificates(configMap *unstructured.Unstructured, module string) int {
	data, _, _ := unstructured.NestedString(configMap.Object, "data", fmt.Sprintf("%s.crt", module))
	return strings.Count(data, "-----BEGIN CERTIFICATE-----")
}
