package test

import (
  	"fmt"
	"testing"
	"time"
	"strings"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terratest/modules/random"
	//"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	err = k8s.SetNestedSliceString(vs.Object, 0, "host", host, "spec", "hosts")
	require.NoError(t, err, "Setting spec.hosts error")
	
	
	resource := schema.GroupVersionResource{Group: "networking.istio.io", Version: "v1beta1", Resource: "virtualservices"}
	createdVS, err := k8s.CreateObject(t, kubectlOptions, vs, namespaceName, resource)
	require.NoError(t, err, "Creating VirtualService error")
	assert.NotNil(t, createdVS, "VirtualService is nil")
	
	
	
	
  
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	if err != nil {
		t.Fatal("external-dns deployment error:", err)
	}
}
