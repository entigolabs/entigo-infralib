package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		if privateCAEnabled(t, e) {
			// The aws-privateca-issuer subchart: the AWSPCAClusterIssuer gets Ready once the controller has verified
			// its IRSA credentials. No certificate is requested from it, the test account has no private CA.
			err := k8s.WaitUntilDeploymentAvailableE(t, c, "aws-privateca-issuer", 60, 10*time.Second)
			require.NoError(t, err, "aws-privateca-issuer deployment error")
			awsPCAClusterIssuerGVR := schema.GroupVersionResource{Group: "awspca.cert-manager.io", Version: "v1beta1", Resource: "awspcaclusterissuers"}
			_, err = waitUntilConditions(t, c, awsPCAClusterIssuerGVR, "", "aws-privateca", []string{"Ready"}, 60, 10*time.Second)
			require.NoError(t, err, "AWSPCAClusterIssuer error")
		}

		if privateCAPolicyEnabled(t, e) {
			// The cert-manager-approver-policy subchart: it alone decides on the requests of the AWS Private CA
			// issuers, so a name inside the policy of test/aws_pri.yaml is approved and one outside of it is denied.
			// Nothing is issued, the CA is a placeholder.
			err := k8s.WaitUntilDeploymentAvailableE(t, c, "cert-manager-approver-policy", 60, 10*time.Second)
			require.NoError(t, err, "cert-manager-approver-policy deployment error")
			policyGVR := schema.GroupVersionResource{Group: "policy.cert-manager.io", Version: "v1alpha1", Resource: "certificaterequestpolicies"}
			_, err = waitUntilConditions(t, c, policyGVR, "", fmt.Sprintf("%s-privateca", c.Namespace), []string{"Ready"}, 60, 10*time.Second)
			require.NoError(t, err, "CertificateRequestPolicy error")
			testPrivateCAPolicy(t, c, "allowed.example.com", "Approved")
			testPrivateCAPolicy(t, c, "denied.example.com", "Denied")
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

// privateCAEnabled reports whether the environment enables the aws-privateca-issuer subchart
// (global.aws.privateCA.enabled).
func privateCAEnabled(t *testing.T, e *env.Environment) bool {
	var value any = env.ModulePlacement(t, e).Module.Inputs
	for _, key := range []string{"global", "aws", "privateCA", "enabled"} {
		values, ok := value.(map[string]any)
		if !ok {
			return false
		}
		value = values[key]
	}
	enabled, _ := value.(bool)
	return enabled
}

// privateCAPolicyEnabled reports whether the environment enables the cert-manager-approver-policy subchart
// (global.aws.privateCA.policy.enabled).
func privateCAPolicyEnabled(t *testing.T, e *env.Environment) bool {
	var value any = env.ModulePlacement(t, e).Module.Inputs
	for _, key := range []string{"global", "aws", "privateCA", "policy", "enabled"} {
		values, ok := value.(map[string]any)
		if !ok {
			return false
		}
		value = values[key]
	}
	enabled, _ := value.(bool)
	return enabled
}

// testPrivateCAPolicy requests a certificate for a random name under the domain from the AWSPCAClusterIssuer and waits
// until its CertificateRequest has the condition, Approved or Denied.
func testPrivateCAPolicy(t *testing.T, c *k8s.Client, domain string, condition string) {
	name := fmt.Sprintf("%s-policy-%s", c.Namespace, strings.ToLower(random.UniqueId()))
	certificateGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificates"}
	certificate := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "cert-manager.io/v1",
		"kind":       "Certificate",
		"metadata":   map[string]any{"name": name, "namespace": c.Namespace},
		"spec": map[string]any{
			"secretName": fmt.Sprintf("%s-tls", name),
			"dnsNames":   []any{fmt.Sprintf("%s.%s", name, domain)},
			"usages":     []any{"server auth", "client auth"},
			"issuerRef":  map[string]any{"group": "awspca.cert-manager.io", "kind": "AWSPCAClusterIssuer", "name": "aws-privateca"},
		},
	}}
	_, err := c.CreateObjectE(certificateGVR, c.Namespace, certificate)
	require.NoError(t, err, "Creating Certificate error")
	defer func() { _ = c.DeleteObjectE(certificateGVR, c.Namespace, name) }()

	requestGVR := schema.GroupVersionResource{Group: "cert-manager.io", Version: "v1", Resource: "certificaterequests"}
	_, err = retry.DoWithRetryE(t, fmt.Sprintf("Wait for the CertificateRequest of %s to be %s", name, condition), 30, 10*time.Second, func() (string, error) {
		requests, err := c.ListObjectsE(requestGVR, c.Namespace, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		for _, request := range requests.Items {
			if request.GetAnnotations()["cert-manager.io/certificate-name"] != name {
				continue
			}
			items, _, _ := unstructured.NestedSlice(request.Object, "status", "conditions")
			status := []string{}
			for _, item := range items {
				if cond, ok := item.(map[string]any); ok {
					if cond["type"] == condition && cond["status"] == "True" {
						return fmt.Sprintf("CertificateRequest %s is %s", request.GetName(), condition), nil
					}
					status = append(status, fmt.Sprintf("%v=%v (%v)", cond["type"], cond["status"], cond["message"]))
				}
			}
			return "", fmt.Errorf("CertificateRequest %s is not %s: %s", request.GetName(), condition, strings.Join(status, ", "))
		}
		return "", fmt.Errorf("no CertificateRequest of Certificate %s", name)
	})
	require.NoError(t, err, "CertificateRequest of %s error", name)
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
