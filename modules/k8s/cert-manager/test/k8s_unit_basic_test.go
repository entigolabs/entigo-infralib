package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sCertManagerDev(t *testing.T) {
	testK8sCertManager(t, "oracle", "dev")
}

func testK8sCertManager(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	// The cert-manager subchart's fullname is the release name itself when the release
	// name already contains "cert-manager", otherwise "<release>-cert-manager".
	baseName := namespaceName
	if !strings.Contains(namespaceName, "cert-manager") {
		baseName = fmt.Sprintf("%s-cert-manager", namespaceName)
	}
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, baseName, 20, 6*time.Second)
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, fmt.Sprintf("%s-cainjector", baseName), 20, 6*time.Second)
	terrak8s.WaitUntilDeploymentAvailable(t, kubectlOptions, fmt.Sprintf("%s-webhook", baseName), 20, 6*time.Second)
}
