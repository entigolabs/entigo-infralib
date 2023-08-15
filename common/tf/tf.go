package tf

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/terraform"
	testStructure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func ApplyTerraform(t *testing.T, bucketName string, awsRegion string, varFile string, workspaceName string) map[string]interface{} {
	key := fmt.Sprintf("%s/terraform.tfstate", os.Getenv("TF_VAR_prefix"))

	rootFolder := ".."
	terraformFolderRelativeToRoot := "test"
	tempTestFolder := testStructure.CopyTerraformFolderToTemp(t, rootFolder, terraformFolderRelativeToRoot)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTestFolder,
		Reconfigure:  true,
		VarFiles:     []string{varFile},
		BackendConfig: map[string]interface{}{
			"bucket": bucketName,
			"key":    key,
			"region": awsRegion,
		},
	})
	terraform.Init(t, terraformOptions)
	terraform.WorkspaceSelectOrNew(t, terraformOptions, workspaceName)

	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer terraform.Destroy(t, terraformOptions)
	}
	terraform.Apply(t, terraformOptions)
	outputs, err := terraform.OutputAllE(t, terraformOptions)
	require.NoError(t, err, "Terraform output error")
	return outputs
}
