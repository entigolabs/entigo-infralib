package test

import (
	"fmt"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

const (
	ConfigurationKind = "configuration.pkg.crossplane.io"
)

func TestK8sPlatformApisAWSBiz(t *testing.T) {
	testK8sPlatformApis(t, "aws", "biz")
}

// func TestK8sPlatformApisAWSPri(t *testing.T) {
// 	testK8sPlatformApis(t, "aws", "pri")
// }

func testK8sPlatformApis(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	argocdNamespace := fmt.Sprintf("argocd-%s", envName)
	argocdOptions := terrak8s.NewKubectlOptions(kubectlOptions.ContextName, kubectlOptions.ConfigPath, argocdNamespace)

	clusterOptions := terrak8s.NewKubectlOptions(kubectlOptions.ContextName, kubectlOptions.ConfigPath, "")

	testPlatformApisZone(t, argocdNamespace, clusterOptions, argocdOptions)
	testPlatformApisPostgresql(t, argocdNamespace, clusterOptions)
}
