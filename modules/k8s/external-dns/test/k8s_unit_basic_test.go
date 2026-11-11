package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/random"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestExternalDns(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilDeploymentAvailableE(t, c, c.Namespace, 10, 6*time.Second), "external-dns deployment error")

		// Oracle has no istio-gateway to hang a VirtualService on and had no
		// record check before either; the deployment check is all there is.
		if e.Cloud == env.CloudOracle {
			return
		}

		// A random hostname under the public domain of the net step.
		domain := publicDomain(t, e)
		host := fmt.Sprintf("%s-%s.%s", strings.ToLower(random.UniqueId()), c.Namespace, domain)
		if e.Cloud == env.CloudAzure {
			testExternalDnsAzure(t, e, c, host, domain)
			return
		}

		vs, err := k8s.ReadObjectFromFile(t, "./templates/virtualService.yaml")
		require.NoError(t, err)
		vs.SetName(fmt.Sprintf("%s-%s", c.Namespace, strings.ToLower(random.UniqueId())))

		err = unstructured.SetNestedStringSlice(vs.Object, []string{host}, "spec", "hosts")
		require.NoError(t, err, "Setting spec.hosts error")

		gw := fmt.Sprintf("istio-gateway-%s/istio-gateway", e.Prefix)
		err = unstructured.SetNestedStringSlice(vs.Object, []string{gw}, "spec", "gateways")
		require.NoError(t, err, "Setting spec.gateways error")

		resource := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "virtualservices"}
		createdVS, err := k8s.CreateObject(t, c, vs, c.Namespace, resource)
		require.NoError(t, err, "Creating VirtualService error")
		assert.NotNil(t, createdVS, "VirtualService is nil")

		switch e.Cloud {
		case env.CloudAWS:
			hostedZoneID := tf.GetStep(t, e, "net").String(t, "route53__pub_zone_id")
			err = aws.WaitUntilRoute53RecordExists(t, e.MustRegion(t), hostedZoneID, host, "A", 12, 15*time.Second)
		case env.CloudGoogle:
			zoneID := tf.GetStep(t, e, "net").String(t, "dns__pub_zone_id")
			err = google.WaitUntilDNSRecordExists(t, e, zoneID, host, "A", 12, 15*time.Second)
		}
		require.NoError(t, err, "Route53Record creation error")
	})
}

// publicDomain is the public domain of the net step: route53 on AWS, dns elsewhere.
func publicDomain(t *testing.T, e *env.Environment) string {
	outputs := tf.GetStep(t, e, "net")
	if e.Cloud == env.CloudAWS {
		return outputs.String(t, "route53__pub_domain")
	}
	return outputs.String(t, "dns__pub_domain")
}

// The istio public Service is ClusterIP on Azure, records come from HTTPRoutes on the AGC gateway
func testExternalDnsAzure(t *testing.T, e *env.Environment, c *k8s.Client, host string, zoneName string) {
	workloadIdentity := schema.GroupVersionResource{Group: "iam.entigo.com", Version: "v1alpha1", Resource: "workloadidentities"}
	_, err := k8s.WaitUntilNamespacedCrossplaneResourceAvailable(t, c, workloadIdentity, c.Namespace, 60, 10*time.Second)
	require.NoError(t, err, "WorkloadIdentity not Ready")

	if e.Prefix == "biz" {
		err = k8s.WaitUntilDeploymentAvailableE(t, c, fmt.Sprintf("%s-external-dns-private", c.Namespace), 10, 6*time.Second)
		require.NoError(t, err, "external-dns-private deployment error")
	}

	routeName := fmt.Sprintf("%s-%s", c.Namespace, strings.ToLower(random.UniqueId()))
	route := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata":   map[string]interface{}{"name": routeName},
		"spec": map[string]interface{}{
			"parentRefs": []interface{}{map[string]interface{}{
				"group":       "gateway.networking.k8s.io",
				"kind":        "Gateway",
				"name":        "external",
				"namespace":   "azure-gateway",
				"sectionName": "https",
			}},
			"hostnames": []interface{}{host},
			"rules": []interface{}{map[string]interface{}{
				"backendRefs": []interface{}{map[string]interface{}{"name": c.Namespace, "port": int64(7979)}},
			}},
		},
	}}
	_, err = k8s.CreateK8SHTTPRoute(t, c, route)
	require.NoError(t, err, "Creating HTTPRoute error")
	defer func() {
		_ = k8s.DeleteK8SHTTPRoute(t, c, routeName)
	}()

	recordName := strings.TrimSuffix(host, "."+zoneName)
	err = azure.WaitUntilDNSRecordExists(t, e, azure.ResourceGroup(t, e), zoneName, recordName, "CNAME", 40, 15*time.Second)
	require.NoError(t, err, "Azure DNS record creation error")
}
