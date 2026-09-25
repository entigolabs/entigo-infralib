package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

func TestK8sTempoAWSBiz(t *testing.T) {
	testK8sTempo(t, "aws", "biz")
}

func TestK8sTempoAWSPri(t *testing.T) {
	testK8sTempo(t, "aws", "pri")
}

// Unlike loki, tempo does not write a predictable seed file to its bucket on startup,
// so this only asserts the deployment comes up. If you add an OTLP-forwarding path
// (e.g. from the alloy module), extend this to push a test span and query it back via
// the tempo API instead — that's the only way to actually confirm the S3 write path
// works end to end.
func testK8sTempo(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, namespaceName := k8s.CheckKubectlConnection(t, cloudName, envName)

	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, namespaceName, 20, 6*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("%s deployment error:", namespaceName), err)
	}
}
