package test

import (
	"fmt"
	"testing"
	"time"
	"strings"
	"os"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/stretchr/testify/require"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sExternalSecretsAWSBiz(t *testing.T) {
	testK8sExternalSecrets(t, "arn:aws:eks:eu-north-1:877483565445:cluster/biz-infra-eks", "biz")
}

func TestK8sExternalSecretsAWSPri(t *testing.T) {
	testK8sExternalSecrets(t, "arn:aws:eks:eu-north-1:877483565445:cluster/pri-infra-eks", "pri")
}

func TestK8sExternalSecretsGoogleBiz(t *testing.T) {
	testK8sExternalSecrets(t, "gke_entigo-infralib2_europe-north1_biz-infra-gke", "biz")
}

func TestK8sExternalSecretsGooglePri(t *testing.T) {
	testK8sExternalSecrets(t, "gke_entigo-infralib2_europe-north1_pri-infra-gke", "pri")
}

func testK8sExternalSecrets(t *testing.T, contextName string, envName string) {
	t.Parallel()
	namespaceName := fmt.Sprintf("external-secrets-%s", envName)
        kubectlOptions := k8s.CheckKubectlConnection(t, contextName, namespaceName)
  

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	if err != nil {
		t.Fatal("external-secrets deployment error:", err)
	}
	appName := strings.TrimSpace(strings.ToLower(os.Getenv("APP_NAME")))
	
	if appName == namespaceName {
		err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-webhook", namespaceName), 10, 12*time.Second)
		if err != nil {
			t.Fatal("external-secrets-webhook deployment error:", err)
		}

		err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, fmt.Sprintf("%s-cert-controller", namespaceName), 10, 12*time.Second)
		if err != nil {
			t.Fatal("external-secrets-cert-controller deployment error:", err)
		}
	}

	_, err = k8s.WaitUntilClusterSecretStoreAvailable(t, kubectlOptions, namespaceName, 10, 6*time.Second)
	require.NoError(t, err, "ClusterSecretStore not available error")
}
