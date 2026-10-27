package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	_ "github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/random"
	"github.com/entigolabs/entigo-infralib-test/retry"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestCertManager(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		for _, deployment := range []string{c.Namespace, fmt.Sprintf("%s-cainjector", c.Namespace), fmt.Sprintf("%s-webhook", c.Namespace)} {
			err := k8s.WaitUntilDeploymentAvailableE(t, c, deployment, 60, 10*time.Second)
			require.NoError(t, err, "%s deployment error", deployment)
		}

		clusterIssuerGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "clusterissuers"}
		_, err := waitUntilConditions(t, c, clusterIssuerGVR, "", "letsencrypt", []string{"Ready"}, 60, 10*time.Second)
		require.NoError(t, err, "ClusterIssuer error")

		// Issue a certificate for a random name under the public domain of the net step.
		domain := tf.GetStep(t, e, "net").String(t, "dns__pub_domain")
		randomName := strings.ToLower(random.UniqueId())
		certificateName := fmt.Sprintf("%s-%s", c.Namespace, randomName)
		secretName := fmt.Sprintf("%s-tls", certificateName)
		certificate := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]any{
				"name":      certificateName,
				"namespace": c.Namespace,
			},
			"spec": map[string]any{
				"secretName": secretName,
				"dnsNames":   []any{fmt.Sprintf("%s.%s", randomName, domain)},
				"issuerRef": map[string]any{
					"kind": "ClusterIssuer",
					"name": "letsencrypt",
				},
			},
		}}
		certificateGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
		_, err = c.CreateObjectE(certificateGVR, c.Namespace, certificate)
		require.NoError(t, err, "Creating Certificate error")
		defer func() {
			_ = c.DeleteObjectE(certificateGVR, c.Namespace, certificateName)
			secretGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
			_ = c.DeleteObjectE(secretGVR, c.Namespace, secretName)
		}()

		_, err = waitUntilConditions(t, c, certificateGVR, c.Namespace, certificateName, []string{"Ready"}, 90, 10*time.Second)
		require.NoError(t, err, "Certificate %s error", certificateName)
	})
}

// waitUntilConditions waits until every named status condition of the object is True.
func waitUntilConditions(t *testing.T, c *k8s.Client, resource schema.GroupVersionResource, namespace string, name string, conditions []string, retries int, sleepBetweenRetries time.Duration) (*unstructured.Unstructured, error) {
	var object *unstructured.Unstructured
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("Wait for %s %s", resource.Resource, name), retries, sleepBetweenRetries, func() (string, error) {
		var err error
		object, err = c.GetObjectE(resource, namespace, name)
		if err != nil {
			return "", err
		}
		status := map[string]string{}
		items, _, _ := unstructured.NestedSlice(object.Object, "status", "conditions")
		for _, item := range items {
			if condition, ok := item.(map[string]any); ok {
				status[fmt.Sprint(condition["type"])] = fmt.Sprint(condition["status"])
			}
		}
		for _, condition := range conditions {
			if status[condition] != "True" {
				return "", fmt.Errorf("%s %s condition %s is %q", resource.Resource, name, condition, status[condition])
			}
		}
		return fmt.Sprintf("%s %s is available", resource.Resource, name), nil
	})
	return object, err
}
