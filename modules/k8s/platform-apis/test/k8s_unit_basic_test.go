package e2e_tests

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
)

func TestK8sPlatformApisAWSBiz(t *testing.T) {
	testK8sPlatformApis(t, "aws", "biz")
}

func TestK8sPlatformApisAWSPri(t *testing.T) {
	testK8sPlatformApis(t, "aws", "pri")
}

func testK8sPlatformApis(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	_, _ = k8s.CheckKubectlConnection(t, cloudName, envName)

}
