 
package test

import (
	"testing"
	"strings"
	"os"
	"github.com/gruntwork-io/terratest/modules/aws"
        "github.com/gruntwork-io/terratest/modules/k8s"
        "github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/terraform"
        "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/davecgh/go-spew/spew"
)


func TestTerraformBasicOne(t *testing.T) {
	
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: ".",
	})
	terraform.WorkspaceSelectOrNew(t, terraformOptions, "one")
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	rootFolder := ".."
        terraformFolderRelativeToRoot := "."
	tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, rootFolder, terraformFolderRelativeToRoot)
        if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
	  spew.Dump(awsRegion)
	  spew.Dump(tempTestFolder)
	}

	assert.Equal(t, 1, len(strings.Split("foo", " ")), "Mock test")
	
	options := &helm.Options{
		SetValues: map[string]string{
			"containerImageRepo":       "nginx",
		},
		KubectlOptions:    k8s.NewKubectlOptions("", "", "kube-system"),
		BuildDependencies: true,
	}
	spew.Dump(options)

}
