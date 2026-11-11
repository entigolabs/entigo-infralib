package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
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

		randomName := strings.ToLower(random.UniqueId())
		certificateName := fmt.Sprintf("%s-%s", c.Namespace, randomName)
		secretName := fmt.Sprintf("%s-tls", certificateName)
		var dnsName string
		var issuerRef map[string]any
		if clusterIssuerCreated(t, e) {
			// The ACME ClusterIssuer: a certificate for a random name under the public domain of the net step, from
			// the route53 module on AWS and the dns module elsewhere.
			clusterIssuerGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "clusterissuers"}
			_, err := waitUntilConditions(t, c, clusterIssuerGVR, "", "letsencrypt", []string{"Ready"}, 60, 10*time.Second)
			require.NoError(t, err, "ClusterIssuer error")
			net := tf.GetStep(t, e, "net")
			domainKey := "dns__pub_domain"
			if net.Has("route53__pub_domain") {
				domainKey = "route53__pub_domain"
			}
			dnsName = fmt.Sprintf("%s.%s", randomName, net.String(t, domainKey))
			issuerRef = map[string]any{"kind": "ClusterIssuer", "name": "letsencrypt"}
		} else {
			// cert-manager alone, for charts that bring their own issuers: a self-signed Issuer in the namespace
			// proves the controller and the webhook work.
			issuerGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "issuers"}
			issuer := &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "cert-manager.io/v1",
				"kind":       "Issuer",
				"metadata":   map[string]any{"name": certificateName, "namespace": c.Namespace},
				"spec":       map[string]any{"selfSigned": map[string]any{}},
			}}
			_, err := c.CreateObjectE(issuerGVR, c.Namespace, issuer)
			require.NoError(t, err, "Creating Issuer error")
			defer func() { _ = c.DeleteObjectE(issuerGVR, c.Namespace, certificateName) }()
			dnsName = fmt.Sprintf("%s.%s.svc", randomName, c.Namespace)
			issuerRef = map[string]any{"kind": "Issuer", "name": certificateName}
		}

		certificate := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]any{
				"name":      certificateName,
				"namespace": c.Namespace,
			},
			"spec": map[string]any{
				"secretName": secretName,
				"dnsNames":   []any{dnsName},
				"issuerRef":  issuerRef,
			},
		}}
		certificateGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
		_, err := c.CreateObjectE(certificateGVR, c.Namespace, certificate)
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

// clusterIssuerCreated reports whether the module creates the ACME ClusterIssuer in the environment: the
// clusterIssuer.create input when set, else the default of the cloud's values (on, except on AWS where the gateways
// use ACM certificates).
func clusterIssuerCreated(t *testing.T, e *env.Environment) bool {
	inputs := env.ModulePlacement(t, e).Module.Inputs
	if clusterIssuer, ok := inputs["clusterIssuer"].(map[string]any); ok {
		if create, ok := clusterIssuer["create"].(bool); ok {
			return create
		}
	}
	return e.Cloud != env.CloudAWS
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
