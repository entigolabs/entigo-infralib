package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
)

func TestK8sGoogleGatewayBiz(t *testing.T) {
	testK8sGoogleGateway(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz")
}

func TestK8sGoogleGatewayPri(t *testing.T) {
	testK8sGoogleGateway(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri")
}

func testK8sGoogleGateway(t *testing.T, contextName, envName string) {
	t.Parallel()

	namespaceName := fmt.Sprintf("google-gateway-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)
	
	//appName := strings.TrimSpace(strings.ToLower(os.Getenv("APP_NAME")))

	_, err := k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-external", namespaceName), 50, 6*time.Second)
	require.NoError(t, err, "google-gateway not available error")
	
	switch envName {
	case "biz":
		_, err = k8s.WaitUntilK8SGatewayAvailable(t, kubectlOptions, fmt.Sprintf("%s-internal", namespaceName), 50, 6*time.Second)
		require.NoError(t, err, "google-gateway not available error")
	}
}
