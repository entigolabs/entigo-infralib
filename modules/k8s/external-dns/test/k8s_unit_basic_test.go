package test

import (
  	"fmt"
	"testing"
	"time"
	"strings"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	terraaws "github.com/gruntwork-io/terratest/modules/aws"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"github.com/entigolabs/entigo-infralib-common/aws"
)

func TestK8sExternalDnsAWSBiz(t *testing.T) {
	testK8sExternalDns(t, "aws", "biz")
}

func TestK8sExternalDnsAWSPri(t *testing.T) {
	testK8sExternalDns(t, "aws", "pri")
}

func TestK8sExternalDnsGoogleBiz(t *testing.T) {
	testK8sExternalDns(t, "google", "biz")
}

func TestK8sExternalDnsGooglePri(t *testing.T) {
	testK8sExternalDns(t, "google", "pri")
}

func testK8sExternalDns(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	_, _, hostName, _ := k8s.GetGatewayConfig(t, cloudName, envName, "external")
	
	vs, err := k8s.ReadObjectFromFile(t, "./templates/virtualservice.yaml")
	require.NoError(t, err)
	vs.SetName(fmt.Sprintf("%s-%s", namespaceName, strings.ToLower(random.UniqueId())))

	host := fmt.Sprintf("%s-%s", strings.ToLower(random.UniqueId()), hostName)
	err = unstructured.SetNestedStringSlice(vs.Object, []string{host}, "spec", "hosts")
	require.NoError(t, err, "Setting spec.hosts error")
	
	gw := fmt.Sprintf("istio-gateway-%s/istio-gateway", envName)
 	err = unstructured.SetNestedStringSlice(vs.Object, []string{gw}, "spec", "gateways")
	require.NoError(t, err, "Setting spec.gateways error")
	
	resource := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "virtualservices"}
	createdVS, err := k8s.CreateObject(t, kubectlOptions, vs, namespaceName, resource)
	require.NoError(t, err, "Creating VirtualService error")
	assert.NotNil(t, createdVS, "VirtualService is nil")
	
	awsRegion := terraaws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	err = aws.WaitUntilAWSRoute53RecordExists(t, hostedZoneID, host, "A", awsRegion, 30, 4*time.Second)
	require.NoError(t, err, "Route53Record creation error")
  
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	if err != nil {
		t.Fatal("external-dns deployment error:", err)
	}
}
