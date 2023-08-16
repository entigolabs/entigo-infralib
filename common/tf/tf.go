package tf

import (
	"errors"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	testStructure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"io/fs"
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
	outputBlocks := getOutputBlocks(t, rootFolder, "outputs.tf")
	createTestTfFile(t, "test.tf", tempTestFolder, variables, outputBlocks)

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

func getOutputBlocks(t *testing.T, rootFolder string, fileName string) []*hcl.Block {
	fullFileName := fmt.Sprintf("%s/%s", rootFolder, fileName)
	if _, err := os.Stat(fullFileName); err != nil && errors.Is(err, fs.ErrNotExist) {
		return []*hcl.Block{}
	}
	outputsFile := ReadTerraformFile(t, fullFileName)
	return getBlocksByType(t, outputsFile, "output")
}

func getBlocksLabels(t *testing.T, file *hcl.File, blockType string) []string {
	blocks := getBlocksByType(t, file, blockType)
	labels := make([]string, 0)
	for _, block := range blocks {
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

func getBlocksByType(t *testing.T, file *hcl.File, blockType string) []*hcl.Block {
	rootSchema := &hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{{Type: blockType, LabelNames: []string{"name"}}}}
	content, _, diags := file.Body.PartialContent(rootSchema)
	if diags.HasErrors() {
		require.NoError(t, diags, "Error reading blocks with type %s", blockType)
	}
	if content == nil {
		return []*hcl.Block{}
	}
	return content.Blocks
}

func createTestTfFile(t *testing.T, fileName string, tempTestFolder string, variables []string, outputBlocks []*hcl.Block) {
	baseTestFile := ReadTerraformFile(t, fmt.Sprintf("%s/%s", baseTestPath, fileName))
	testFile := hclwrite.NewEmptyFile()
	testFileBody := testFile.Body()
	testModule := testFileBody.AppendNewBlock("module", []string{"test"})
	testModuleBody := testModule.Body()
	testModuleBody.SetAttributeValue("source", cty.StringVal("../"))
	addVariables(variables, testModuleBody)
	addOutputs(t, outputBlocks, testFileBody)

	WriteTerraformFile(t, tempTestFolder, fileName, append(baseTestFile.Bytes, testFile.Bytes()...))
}

func addVariables(variables []string, testModuleBody *hclwrite.Body) {
	for _, variable := range variables {
		if variable == "" {
			continue
		}
		testModuleBody.SetAttributeRaw(variable, getTokens("var."+variable))
	}
}

func addOutputs(t *testing.T, outputBlocks []*hcl.Block, testFileBody *hclwrite.Body) {
	for _, block := range outputBlocks {
		if (len(block.Labels)) == 0 {
			continue
		}
		label := block.Labels[0]
		if label == "" {
			continue
		}
		attr, diags := block.Body.JustAttributes()
		if diags.HasErrors() {
			logger.Logf(t, "Error reading output %s attributes: %s", label, diags.Error())
			continue
		}
		output := testFileBody.AppendNewBlock("output", []string{label})
		output.Body().SetAttributeRaw("value", getTokens("module.test."+label))
		addOutputSensitiveAttribute(attr, output.Body())
	}
}

func addOutputSensitiveAttribute(attr hcl.Attributes, body *hclwrite.Body) {
	for name, attr := range attr {
		if name != "sensitive" {
			continue
		}
		value, hclDiags := attr.Expr.Value(nil)
		if hclDiags != nil && hclDiags.HasErrors() {
			value = cty.BoolVal(true)
		}
		body.SetAttributeValue(name, value)
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
