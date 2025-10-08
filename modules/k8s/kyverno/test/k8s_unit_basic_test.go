package test

import (
	"testing"
	"time"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/entigolabs/entigo-infralib-common/k8s"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
)

func TestK8sKyvernoAWSBiz(t *testing.T) {
	testK8sKyverno(t, "aws", "biz")
}

func TestK8sKyvernoAWSPri(t *testing.T) {
	testK8sKyverno(t, "aws", "pri")
}

func TestK8sKyvernoGoogleBiz(t *testing.T) {
	testK8sKyverno(t, "google", "biz")
}

func TestK8sKyvernoGooglePri(t *testing.T) {
	testK8sKyverno(t, "google", "pri")
}

func testK8sKyverno(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)
	
	err := terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "kyverno-admission-controller", 10, 6*time.Second)
	if err != nil {
		t.Fatal("kyverno-admission-controller deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "kyverno-background-controller", 10, 6*time.Second)
	if err != nil {
		t.Fatal("kyverno-background-controller deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "kyverno-cleanup-controller", 10, 6*time.Second)
	if err != nil {
		t.Fatal("kyverno-cleanup-controller deployment error:", err)
	}
	err = terrak8s.WaitUntilDeploymentAvailableE(t, kubectlOptions, "kyverno-reports-controller", 10, 6*time.Second)
	if err != nil {
		t.Fatal("kyverno-reports-controller deployment error:", err)
	}

  

}
