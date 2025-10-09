package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/k8s"
)

func TestK8sMetricsServerAWSBiz(t *testing.T) {
	testK8sMetricsServer(t, "aws", "biz")
}

func TestK8sMetricsServerAWSPri(t *testing.T) {
	testK8sMetricsServer(t, "aws", "pri")
}

func testK8sMetricsServer(t *testing.T, cloudName string, envName string) {
  	t.Parallel()
	_, _ := k8s.CheckKubectlConnection(t, cloudName, envName)


}
