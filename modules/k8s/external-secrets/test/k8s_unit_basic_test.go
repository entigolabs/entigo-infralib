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
	testK8sExternalSecrets(t, "aws", "biz")
}

func TestK8sExternalSecretsAWSPri(t *testing.T) {
	testK8sExternalSecrets(t, "aws", "pri")
}

func TestK8sExternalSecretsGoogleBiz(t *testing.T) {
	testK8sExternalSecrets(t, "google", "biz")
}

func TestK8sExternalSecretsGooglePri(t *testing.T) {
	testK8sExternalSecrets(t, "google", "pri")
}

func testK8sExternalSecrets(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)
  

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
