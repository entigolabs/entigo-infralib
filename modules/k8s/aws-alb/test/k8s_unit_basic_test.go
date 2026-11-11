package test

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	_ "github.com/entigolabs/entigo-infralib-test/aws" // tf output reader for the net step
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
	"github.com/entigolabs/entigo-infralib-test/random"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

// GatewayClass and Gateways are provisioned by the module itself,
// the test only verifies their health and attaches HTTPRoutes.
const gatewayClassName = "alb"

func TestAwsAlbGatewayApi(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"aws_biz": func(t *testing.T, e *env.Environment) {
			// access logs are off by default, the service gateway sets sourceRanges as a list
			testGatewayApi(t, e, []string{"external", "service"}, []string{"external", "service"},
				map[string][]string{"service": {"0.0.0.0/0", "10.0.0.0/8"}}, false)
		},
		"aws_pri": func(t *testing.T, e *env.Environment) {
			// service gateway is disabled in pri, access logs are off for the internal gateway,
			// the external gateway sets sourceRanges as a quoted comma-separated string
			// pri sends the ALB Server header (serverHeader: true), biz uses the default
			testGatewayApi(t, e, []string{"external", "internal"}, []string{"internal"},
				map[string][]string{"external": {"0.0.0.0/0", "10.0.0.0/8"}}, true)
		},
	})
}

// The public domain of the net step is the base for the randomized test hostnames.
func publicDomain(t *testing.T, e *env.Environment) string {
	return tf.GetStep(t, e, "net").String(t, "route53__pub_domain")
}

func testGatewayApi(t *testing.T, e *env.Environment, gatewayNames []string, accessLogsDisabled []string, sourceRanges map[string][]string, serverHeader bool) {
	c := k8s.Connect(t, e)
	hostName := publicDomain(t, e)

	k8s.WaitUntilDeploymentAvailable(t, c, fmt.Sprintf("%s-aws-load-balancer-controller", c.Namespace), 10, 6*time.Second)

	// Health check of the module-provisioned GatewayClass: Accepted must be True
	acceptedStatus, err := k8s.RunKubectlAndGetOutputE(t, c,
		"get", "gatewayclass", gatewayClassName,
		"-o", `jsonpath={.status.conditions[?(@.type=="Accepted")].status}`)
	require.NoError(t, err, "Getting GatewayClass %s error", gatewayClassName)
	require.Equal(t, "True", acceptedStatus, "GatewayClass %s is not Accepted", gatewayClassName)

	// One shared backend workload for all gateway checks
	resourceName := fmt.Sprintf("%s-%s", c.Namespace, strings.ToLower(random.UniqueId()))

	deployment, err := k8s.ReadObjectFromFile(t, "./templates/deployment.yaml")
	require.NoError(t, err)
	deployment.SetName(resourceName)
	_, err = k8s.CreateK8SDeployment(t, c, deployment)
	require.NoError(t, err, "Creating Deployment error")

	service, err := k8s.ReadObjectFromFile(t, "./templates/service.yaml")
	require.NoError(t, err)
	service.SetName(resourceName)
	_, err = k8s.CreateK8SService(t, c, service)
	require.NoError(t, err, "Creating Service error")

	k8s.WaitUntilDeploymentAvailable(t, c, resourceName, 20, 6*time.Second)

	for _, gatewayName := range gatewayNames {
		testGatewayAccessLogs(t, c, gatewayName, !slices.Contains(accessLogsDisabled, gatewayName))
		testGatewaySourceRanges(t, c, gatewayName, sourceRanges[gatewayName])
		testGatewaySslPolicy(t, c, gatewayName)
	}

	// Every enabled gateway gets its own HTTPRoute with a unique hostname
	for _, gatewayName := range gatewayNames {
		testGatewayHTTPRoute(t, c, gatewayName, hostName, resourceName, serverHeader)
	}

	err = k8s.DeleteK8SService(t, c, resourceName)
	require.NoError(t, err, "Deleting Service error")
	err = k8s.DeleteK8SDeployment(t, c, resourceName)
	require.NoError(t, err, "Deleting Deployment error")
}

func testGatewayAccessLogs(t *testing.T, c *k8s.Client, gatewayName string, enabled bool) {
	getAttribute := func(key string) string {
		value, err := k8s.RunKubectlAndGetOutputE(t, c,
			"get", "loadbalancerconfigurations.gateway.k8s.aws", gatewayName,
			"-o", fmt.Sprintf(`jsonpath={.spec.loadBalancerAttributes[?(@.key=="%s")].value}`, key))
		require.NoError(t, err, "Getting LoadBalancerConfiguration %s error", gatewayName)
		return value
	}

	require.Equal(t, fmt.Sprintf("%t", enabled), getAttribute("access_logs.s3.enabled"), "Gateway %s access_logs.s3.enabled", gatewayName)
	if !enabled {
		return
	}
	require.Equal(t, gatewayName, getAttribute("access_logs.s3.prefix"), "Gateway %s access_logs.s3.prefix", gatewayName)
	bucket := getAttribute("access_logs.s3.bucket")
	require.NotEmpty(t, bucket, "Gateway %s access_logs.s3.bucket is empty", gatewayName)

	// The test values set no bucket, so the gateway uses a bucket created by the module
	ready, err := k8s.RunKubectlAndGetOutputE(t, c,
		"get", "buckets.s3.aws.upbound.io", bucket,
		"-o", `jsonpath={.status.conditions[?(@.type=="Ready")].status}`)
	require.NoError(t, err, "Getting access logs Bucket %s error", bucket)
	require.Equal(t, "True", ready, "Access logs Bucket %s is not Ready", bucket)
}

// The test values set no sslPolicy, so every gateway gets the TLS 1.3 only default
func testGatewaySslPolicy(t *testing.T, c *k8s.Client, gatewayName string) {
	value, err := k8s.RunKubectlAndGetOutputE(t, c,
		"get", "loadbalancerconfigurations.gateway.k8s.aws", gatewayName,
		"-o", `jsonpath={.spec.listenerConfigurations[?(@.protocolPort=="HTTPS:443")].sslPolicy}`)
	require.NoError(t, err, "Getting LoadBalancerConfiguration %s error", gatewayName)
	require.Equal(t, "ELBSecurityPolicy-TLS13-1-3-2021-06", value, "Gateway %s sslPolicy", gatewayName)
}

// Gateways without sourceRanges in the test values must not set any, the ALB then allows 0.0.0.0/0
func testGatewaySourceRanges(t *testing.T, c *k8s.Client, gatewayName string, expected []string) {
	value, err := k8s.RunKubectlAndGetOutputE(t, c,
		"get", "loadbalancerconfigurations.gateway.k8s.aws", gatewayName,
		"-o", `jsonpath={.spec.sourceRanges[*]}`)
	require.NoError(t, err, "Getting LoadBalancerConfiguration %s error", gatewayName)
	if len(expected) == 0 {
		require.Empty(t, value, "Gateway %s has unexpected sourceRanges", gatewayName)
		return
	}
	require.Equal(t, expected, strings.Fields(value), "Gateway %s sourceRanges", gatewayName)
}

func testGatewayHTTPRoute(t *testing.T, c *k8s.Client, gatewayName string, hostName string, backendName string, serverHeader bool) {
	// Health check of the module-provisioned Gateway, the helper waits for Programmed
	existingGateway, err := k8s.WaitUntilK8SGatewayAvailable(t, c, gatewayName, 60, 5*time.Second)
	require.NoError(t, err, "Gateway %s availability error", gatewayName)

	// The Gateway must reference the expected GatewayClass
	className, found, err := unstructured.NestedString(existingGateway.Object, "spec", "gatewayClassName")
	require.NoError(t, err)
	require.True(t, found, "Gateway %s spec.gatewayClassName not found", gatewayName)
	require.Equal(t, gatewayClassName, className, "Gateway %s does not use GatewayClass %s", gatewayName, gatewayClassName)

	gatewayAddress := k8s.GetK8SGatewayAddress(existingGateway)
	require.NotEmpty(t, gatewayAddress, "Gateway %s address is empty", gatewayName)

	// Randomized hostname, same mechanism as the ingress test
	hostname := fmt.Sprintf("%s-%s", strings.ToLower(random.UniqueId()), hostName)
	routeName := fmt.Sprintf("%s-%s", backendName, gatewayName)

	httpRoute, err := k8s.ReadObjectFromFile(t, "./templates/httpRoute.yaml")
	require.NoError(t, err)
	httpRoute.SetName(routeName)
	err = k8s.SetNestedSliceString(httpRoute.Object, 0, "name", gatewayName, "spec", "parentRefs")
	require.NoError(t, err, "Setting HTTPRoute parentRef name error")
	err = k8s.SetNestedSliceString(httpRoute.Object, 0, "namespace", c.Namespace, "spec", "parentRefs")
	require.NoError(t, err, "Setting HTTPRoute parentRef namespace error")
	err = unstructured.SetNestedStringSlice(httpRoute.Object, []string{hostname}, "spec", "hostnames")
	require.NoError(t, err, "Setting HTTPRoute hostnames error")
	rules, found, err := unstructured.NestedSlice(httpRoute.Object, "spec", "rules")
	require.NoError(t, err)
	require.True(t, found, "HTTPRoute spec.rules not found")
	backendRefs, ok := rules[0].(map[string]interface{})["backendRefs"].([]interface{})
	require.True(t, ok, "HTTPRoute spec.rules[0].backendRefs not found")
	backendRefs[0].(map[string]interface{})["name"] = backendName
	rules[0].(map[string]interface{})["backendRefs"] = backendRefs
	err = unstructured.SetNestedSlice(httpRoute.Object, rules, "spec", "rules")
	require.NoError(t, err, "Setting HTTPRoute backendRef name error")
	createdHTTPRoute, err := k8s.CreateK8SHTTPRoute(t, c, httpRoute)
	require.NoError(t, err, "Creating HTTPRoute for gateway %s error", gatewayName)
	assert.NotNil(t, createdHTTPRoute)

	_, err = k8s.WaitUntilK8SHTTPRouteAvailable(t, c, routeName, 60, 5*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SHTTPRoute(t, c, routeName)
	}
	require.NoError(t, err, "HTTPRoute availability error for gateway %s", gatewayName)

	targetURL := fmt.Sprintf("http://%s/", hostname)
	err = k8s.WaitUntilHostnameAvailableWithAddress(t, c, gatewayAddress, targetURL, "200", 60, 5*time.Second)
	if err != nil {
		_ = k8s.DeleteK8SHTTPRoute(t, c, routeName)
	}
	require.NoError(t, err, "HTTPRoute HTTP test error for gateway %s", gatewayName)

	// The HTTP listener exists while the route is attached
	err = testGatewayServerHeader(t, c, gatewayAddress, fmt.Sprintf("%s-noroute-%s", strings.ToLower(random.UniqueId()), hostName), serverHeader)
	if err != nil {
		_ = k8s.DeleteK8SHTTPRoute(t, c, routeName)
	}
	require.NoError(t, err, "Server header test error for gateway %s", gatewayName)

	err = k8s.DeleteK8SHTTPRoute(t, c, routeName)
	require.NoError(t, err, "Deleting HTTPRoute error for gateway %s", gatewayName)
	err = k8s.WaitUntilK8SHTTPRouteDeleted(t, c, routeName, 40, 2*time.Second)
	require.NoError(t, err, "HTTPRoute didn't get deleted for gateway %s", gatewayName)
}

// Requests a host without a route, so the ALB answers itself, and checks its Server: awselb header
func testGatewayServerHeader(t *testing.T, c *k8s.Client, gatewayAddress string, hostname string, expected bool) error {
	job, err := k8s.ReadObjectFromFile(t, "./templates/serverHeaderJob.yaml")
	if err != nil {
		return err
	}
	jobName := fmt.Sprintf("%s-server-header-%s", c.Namespace, strings.ToLower(random.UniqueId()))
	job.SetName(jobName)
	containers, _, err := unstructured.NestedSlice(job.Object, "spec", "template", "spec", "containers")
	if err != nil {
		return err
	}
	containers[0].(map[string]interface{})["env"] = []interface{}{
		map[string]interface{}{"name": "TARGET_DOMAIN", "value": hostname},
		map[string]interface{}{"name": "TARGET_IP", "value": gatewayAddress},
		map[string]interface{}{"name": "EXPECT_SERVER_HEADER", "value": fmt.Sprintf("%t", expected)},
	}
	if err = unstructured.SetNestedSlice(job.Object, containers, "spec", "template", "spec", "containers"); err != nil {
		return err
	}

	if _, err = k8s.CreateObject(t, c, job, c.Namespace, k8s.Jobs); err != nil {
		return err
	}
	err = k8s.WaitUntilJobSucceedE(t, c, jobName, 60, 6*time.Second)
	logs, _ := k8s.RunKubectlAndGetOutputE(t, c, "logs", "job/"+jobName)
	t.Logf("Server header check for %s:\n%s", gatewayAddress, logs)
	_, _ = k8s.RunKubectlAndGetOutputE(t, c, "delete", "job", jobName, "--wait=false")
	return err
}

func TestAwsAlb(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)
		hostName := publicDomain(t, e)

		k8s.WaitUntilDeploymentAvailable(t, c, fmt.Sprintf("%s-aws-load-balancer-controller", c.Namespace), 10, 6*time.Second)
		k8s.WaitUntilServiceAvailable(t, c, "aws-load-balancer-webhook-service", 60, 1*time.Second)
		time.Sleep(5 * time.Second)

		ingress, err := k8s.ReadObjectFromFile(t, "./templates/ingress.yaml")
		require.NoError(t, err)
		ingress.SetName(fmt.Sprintf("%s-%s", c.Namespace, strings.ToLower(random.UniqueId())))
		ingressClass := "alb"
		err = unstructured.SetNestedField(ingress.Object, ingressClass, "spec", "ingressClassName")
		require.NoError(t, err, "Setting ingressClassName error")
		annotations := ingress.GetAnnotations()
		annotations["alb.ingress.kubernetes.io/group.name"] = "aws-load-balancer"
		ingress.SetAnnotations(annotations)
		host := fmt.Sprintf("%s-%s", strings.ToLower(random.UniqueId()), hostName)
		err = k8s.SetNestedSliceString(ingress.Object, 0, "host", host, "spec", "rules")
		require.NoError(t, err, "Setting host error")
		createdIngress, err := k8s.CreateK8SIngress(t, c, ingress)
		require.NoError(t, err, "Creating ingress error")
		assert.NotNil(t, createdIngress, "Ingress is nil")

		_, err = k8s.WaitUntilK8SIngressAvailable(t, c, createdIngress.GetName(), 40, 2*time.Second)
		if err != nil {
			_ = k8s.DeleteK8SIngress(t, c, ingress.GetName()) // Try to delete ingress
		}
		require.NoError(t, err, "Ingress availability error")

		err = k8s.DeleteK8SIngress(t, c, ingress.GetName())
		require.NoError(t, err, "Deleting ingress error")

		err = k8s.WaitUntilK8SIngressDeleted(t, c, ingress.GetName(), 40, 2*time.Second)
		require.NoError(t, err, "Ingress didn't get deleted")
	})
}
