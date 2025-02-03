package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sMimirAWSBiz(t *testing.T) {
	testK8sMimir(t, "aws", "biz")
}

func TestK8sMimirAWSPri(t *testing.T) {
	testK8sMimir(t, "aws", "pri")
}

func TestK8sMimirGoogleBiz(t *testing.T) {
	testK8sMimir(t, "google", "biz")
}

func TestK8sMimirGooglePri(t *testing.T) {
	testK8sMimir(t, "google", "pri")
}

func testK8sMimir(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	gatewayName, gatewayNamespace, hostName, retries := k8s.GetGatewayConfig(t, cloudName, envName, "default")
	
	if cloudName == "aws" {
		gatewayName = "mimir-gateway"
	}

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-gateway", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-gateway deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-distributor", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-distributor deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-querier", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-querier deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-query-frontend", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-query-frontend deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "mimir-ruler", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-ruler deployment error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-compactor-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-compactor-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-ingester-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-ingester-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "mimir-store-gateway-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("mimir-store-gateway-0 pod error:", err)
	}
	
	switch cloudName {
	  case "aws":
	    err = aws.WaitUntilBucketFileAvailable(t, fmt.Sprintf("%s-%s-877483565445-eu-north-1", envName, namespaceName), "__mimir_cluster/mimir_cluster_seed.json", 20, 6*time.Second)
	    if err != nil {
		    t.Fatal("File not found in AWS bucket:", err)
	    }
	  case "google":
	    err = google.WaitUntilBucketFileAvailable(t, fmt.Sprintf("%s-%s-metrics", envName, namespaceName), "blocks/__mimir_cluster/mimir_cluster_seed.json", 20, 6*time.Second)
	    if err != nil {
		    t.Fatal("File not found in Google bucket:", err)
	    }
	}

	successResponseCode := "200"
	targetURL := fmt.Sprintf("https://%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, fmt.Sprintf("%s ingress/gateway test error", namespaceName))
}
