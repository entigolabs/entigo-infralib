package tf

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/terraform"
	testStructure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"os"
	"testing"
)

const baseTestPath = "/common/tf"

func ApplyTerraform(t *testing.T, bucketName string, awsRegion string, varFile string, workspaceName string) (map[string]interface{}, func()) {
	key := fmt.Sprintf("%s/terraform.tfstate", os.Getenv("TF_VAR_prefix"))

	rootFolder := ".."
	terraformFolderRelativeToRoot := "test"
	tempTestFolder := testStructure.CopyTerraformFolderToTemp(t, rootFolder, terraformFolderRelativeToRoot)

	variables := copyVariablesFile(t, rootFolder, "variables.tf", tempTestFolder)
	outputLabels := getOutputLabels(t, rootFolder, "outputs.tf")
	createTestTfFile(t, "test.tf", tempTestFolder, variables, outputLabels)

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

	var destroy = func() {
		// Do nothing
	}
	if os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		destroy = func() { terraform.Destroy(t, terraformOptions) }
	}
	terraform.Apply(t, terraformOptions)
	outputs, err := terraform.OutputAllE(t, terraformOptions)
	if err != nil && os.Getenv("ENTIGO_INFRALIB_DESTROY") == "true" {
		defer terraform.Destroy(t, terraformOptions)
	}
	require.NoError(t, err, "Terraform output error")
	return outputs, destroy
}

func copyVariablesFile(t *testing.T, rootFolder string, fileName string, tempTestFolder string) []string {
	variablesFile := ReadTerraformFile(t, fmt.Sprintf("%s/%s", rootFolder, fileName))
	WriteTerraformFile(t, tempTestFolder, fileName, variablesFile.Bytes)
	return getBlocksLabels(t, variablesFile, "variable")
}

func getOutputLabels(t *testing.T, rootFolder string, fileName string) []string {
	outputsFile := ReadTerraformFile(t, fmt.Sprintf("%s/%s", rootFolder, fileName))
	return getBlocksLabels(t, outputsFile, "output")
}

func getBlocksLabels(t *testing.T, file *hcl.File, blockType string) []string {
	rootSchema := &hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{{Type: blockType, LabelNames: []string{"name"}}}}
	content, _, diags := file.Body.PartialContent(rootSchema)
	if diags.HasErrors() {
		require.NoError(t, diags, "Error reading blocks with type %s", blockType)
	}
	if content == nil || len(content.Blocks) == 0 {
		return []string{}
	}
	labels := make([]string, 0)
	for _, block := range content.Blocks {
		if (len(block.Labels)) == 0 {
			continue
		}
		label := block.Labels[0]
		if label == "" {
			continue
		}
		labels = append(labels, label)
	}
	return labels
}

func createTestTfFile(t *testing.T, fileName string, tempTestFolder string, variables []string, outputLabels []string) {
	baseTestFile := ReadTerraformFile(t, fmt.Sprintf("%s/%s", baseTestPath, fileName))
	testFile := hclwrite.NewEmptyFile()
	testFileBody := testFile.Body()
	testModule := testFileBody.AppendNewBlock("module", []string{"test"})
	testModuleBody := testModule.Body()
	testModuleBody.SetAttributeValue("source", cty.StringVal("../"))
	addVariables(variables, testModuleBody)
	addOutputs(outputLabels, testFileBody)

	WriteTerraformFile(t, tempTestFolder, fileName, append(baseTestFile.Bytes, testFile.Bytes()...))
}

func addVariables(variables []string, testModuleBody *hclwrite.Body) {
	if len(variables) == 0 {
		return
	}
	for _, variable := range variables {
		if variable == "" {
			continue
		}
		testModuleBody.SetAttributeRaw(variable, getTokens("var."+variable))
	}
}

func addOutputs(outputLabels []string, testFileBody *hclwrite.Body) {
	if len(outputLabels) == 0 {
		return
	}
	for _, outputLabel := range outputLabels {
		if outputLabel == "" {
			continue
		}
		output := testFileBody.AppendNewBlock("output", []string{outputLabel})
		output.Body().SetAttributeRaw("value", getTokens("module.test."+outputLabel))
	}
}

func getTokens(value string) hclwrite.Tokens {
	return hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte(value),
		},
	}
}

func ReadTerraformFile(t *testing.T, fileName string) *hcl.File {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(fileName)
	if diags.HasErrors() {
		require.NoError(t, diags, "Error reading %s", fileName)
	}
	return file
}

func WriteTerraformFile(t *testing.T, path string, fileName string, bytes []byte) {
	fullFileName := fmt.Sprintf("%s/%s", path, fileName)
	tfFile, err := os.Create(fullFileName)
	require.NoError(t, err, "Error creating %s", fullFileName)
	_, err = tfFile.Write(hclwrite.Format(bytes))
	require.NoError(t, err, "%s file write error", fullFileName)
}
