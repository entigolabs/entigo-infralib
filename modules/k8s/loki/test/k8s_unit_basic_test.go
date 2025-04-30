package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sLokiAWSBiz(t *testing.T) {
	testK8sLoki(t, "aws", "biz")
}

func TestK8sLokiAWSPri(t *testing.T) {
	testK8sLoki(t, "aws", "pri")
}

func TestK8sLokiGoogleBiz(t *testing.T) {
	testK8sLoki(t, "google", "biz")
}

func TestK8sLokiGooglePri(t *testing.T) {
	testK8sLoki(t, "google", "pri")
}

func testK8sLoki(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	gatewayName, gatewayNamespace, hostName, retries := k8s.GetGatewayConfig(t, cloudName, envName, "default")

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "loki-read", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-read deployment error:", err)
	}

	lokiGatewayName := fmt.Sprintf("%s-gateway", namespaceName)
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, lokiGatewayName, 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-gateway deployment error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "loki-write-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-write-0 pod error:", err)
	}
	err = terrak8s.WaitUntilPodAvailableE(t, kubectlOptions, "loki-backend-0", 20, 6*time.Second)
	if err != nil {
		t.Fatal("loki-backend-0 pod error:", err)
	}

	switch cloudName {
	case "aws":
		gatewayName = fmt.Sprintf("%s-gateway", gatewayName)

		err = aws.WaitUntilBucketFileAvailable(t, fmt.Sprintf("%s-%s-877483565445-eu-north-1", envName, namespaceName), "loki_cluster_seed.json", 20, 6*time.Second)
		if err != nil {
			t.Fatal("File not found in AWS bucket:", err)
		}
	case "google":
		err = google.WaitUntilBucketFileAvailable(t, fmt.Sprintf("%s-%s-logs", envName, namespaceName), "loki_cluster_seed.json", 20, 6*time.Second)
		if err != nil {
			t.Fatal("File not found in Google bucket:", err)
		}
	}

	successResponseCode := "200"
	targetURL := fmt.Sprintf("https://%s", hostName)
	err = k8s.WaitUntilHostnameAvailable(t, kubectlOptions, retries, 6*time.Second, gatewayName, gatewayNamespace, namespaceName, targetURL, successResponseCode, cloudName)
	require.NoError(t, err, fmt.Sprintf("%s ingress/gateway test error", namespaceName))
}
