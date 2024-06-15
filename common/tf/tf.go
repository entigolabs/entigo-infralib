package tf

import (
	"errors"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/terraform"
	testStructure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"io/fs"
	"os"
	"testing"
)

const rootFolder = ".."
const terraformFolderRelativeToRoot = "test"
const providersPath = "/providers"
const testProvidersPath = "./providers"

type ProviderType string

const (
	AWS    ProviderType = "aws"
	GCloud ProviderType = "gcloud"
)

func InitAWSTerraform(t *testing.T, bucketName string, awsRegion string, varFile string, vars map[string]interface{}) *terraform.Options {
	return InitTerraform(t, bucketName, awsRegion, varFile, vars, AWS)
}

func InitGCloudTerraform(t *testing.T, bucketName string, gcloudRegion string, varFile string, vars map[string]interface{}) *terraform.Options {
	return InitTerraform(t, bucketName, gcloudRegion, varFile, vars, GCloud)
}

func InitTerraform(t *testing.T, bucketName string, awsRegion string, varFile string, vars map[string]interface{}, providerType ProviderType) *terraform.Options {
	key := fmt.Sprintf("%s/terraform.tfstate", os.Getenv("TF_VAR_prefix"))

	tempTestFolder := testStructure.CopyTerraformFolderToTemp(t, rootFolder, terraformFolderRelativeToRoot)

	variables := copyVariablesFile(t, "variables.tf", tempTestFolder)
	outputBlocks := getOutputBlocks(t, "outputs.tf")
	versionsAttributes := getProviderVersions(t, "versions.tf")

	createTestTfFile(t, "test.tf", tempTestFolder, variables, outputBlocks, versionsAttributes, providerType)
	backendConfig := map[string]interface{}{
		"bucket": bucketName,
	}
	if providerType == GCloud {
		backendConfig["prefix"] = key
	} else {
		backendConfig["key"] = key
		backendConfig["region"] = awsRegion
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir:  tempTestFolder,
		Reconfigure:   true,
		Vars:          vars,
		VarFiles:      []string{varFile},
		BackendConfig: backendConfig,
	})
	terraform.Init(t, terraformOptions)
	return terraformOptions
}

func ApplyTerraform(t *testing.T, workspaceName string, terraformOptions *terraform.Options) (map[string]interface{}, func()) {

	_, err := terraform.WorkspaceSelectOrNewE(t, terraformOptions, workspaceName)
	if err != nil {
		terraform.WorkspaceSelectOrNew(t, terraformOptions, workspaceName) // Retry once
	}

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

func copyVariablesFile(t *testing.T, fileName string, tempTestFolder string) []string {
	variablesFile := ReadTerraformFile(t, fmt.Sprintf("%s/%s", rootFolder, fileName))
	WriteTerraformFile(t, tempTestFolder, fileName, variablesFile.Bytes())
	return getBlocksLabels(variablesFile, "variable")
}

func getOutputBlocks(t *testing.T, fileName string) []*hclwrite.Block {
	fullFileName := fmt.Sprintf("%s/%s", rootFolder, fileName)
	if _, err := os.Stat(fullFileName); err != nil && errors.Is(err, fs.ErrNotExist) {
		return []*hclwrite.Block{}
	}
	outputsFile := ReadTerraformFile(t, fullFileName)
	return getBlocksByType(outputsFile, "output")
}

func getProviderVersions(t *testing.T, file string) map[string]*hclwrite.Attribute {
	fullFileName := fmt.Sprintf("%s/%s", rootFolder, file)
	if _, err := os.Stat(fullFileName); err != nil && errors.Is(err, fs.ErrNotExist) {
		require.NoError(t, err, "Error reading %s", fullFileName)
	}
	versionsFile := ReadTerraformFile(t, fullFileName)
	return getRequiredProvidersBlock(t, versionsFile).Body().Attributes()
}

func getBlocksLabels(file *hclwrite.File, blockType string) []string {
	blocks := getBlocksByType(file, blockType)
	labels := make([]string, 0)
	for _, block := range blocks {
		if (len(block.Labels())) == 0 {
			continue
		}
		label := block.Labels()[0]
		if label == "" {
			continue
		}
		labels = append(labels, label)
	}
	return labels
}

func getBlocksByType(file *hclwrite.File, blockType string) []*hclwrite.Block {
	blocks := make([]*hclwrite.Block, 0)
	for _, block := range file.Body().Blocks() {
		if block == nil || block.Type() != blockType {
			continue
		}
		blocks = append(blocks, block)
	}
	return blocks
}

func getRequiredProvidersBlock(t *testing.T, file *hclwrite.File) *hclwrite.Block {
	terraformBlock := file.Body().FirstMatchingBlock("terraform", []string{})
	if terraformBlock == nil {
		require.NoError(t, errors.New("terraform block not found"), "Error parsing tf file")
	}
	providersBlock := terraformBlock.Body().FirstMatchingBlock("required_providers", []string{})
	if providersBlock == nil {
		require.NoError(t, errors.New("required_providers block not found"), "Error parsing tf file")
	}
	return providersBlock
}

func createTestTfFile(t *testing.T, fileName string, tempTestFolder string, variables []string,
	outputBlocks []*hclwrite.Block, versionsAttributes map[string]*hclwrite.Attribute, providerType ProviderType) {

	testFile := ReadTerraformFile(t, fmt.Sprintf("%s/%s", providersPath, "base.tf"))
	testFileBody := testFile.Body()
	modifyBackendType(t, testFileBody, providerType)
	providersBlock := getRequiredProvidersBlock(t, testFile)
	for name, attribute := range versionsAttributes {
		providersBlock.Body().SetAttributeRaw(name, attribute.Expr().BuildTokens(nil))
		providerBlocks := getProviderBlocks(t, name, providerType)
		for _, providerBlock := range providerBlocks {
			testFileBody.AppendBlock(providerBlock)
		}
	}
	testModule := testFileBody.AppendNewBlock("module", []string{"test"})
	testModuleBody := testModule.Body()
	testModuleBody.SetAttributeValue("source", cty.StringVal("../"))
	addVariables(variables, testModuleBody)
	addOutputs(outputBlocks, testFileBody)
	WriteTerraformFile(t, tempTestFolder, fileName, testFile.Bytes())
}

func modifyBackendType(t *testing.T, body *hclwrite.Body, providerType ProviderType) {
	if providerType != GCloud {
		return
	}
	terraformBlock := body.FirstMatchingBlock("terraform", []string{})
	require.NotNil(t, terraformBlock, "terraform block not found")
	backendBlock := terraformBlock.Body().FirstMatchingBlock("backend", []string{"s3"})
	require.NotNil(t, backendBlock, "backend block not found")
	backendBlock.SetLabels([]string{"gcs"})
}

func getProviderBlocks(t *testing.T, providerName string, providerType ProviderType) []*hclwrite.Block {
	if providerName == "helm" {
		switch providerType {
		case "aws":
			providerName = "helm_aws"
		case "gcloud":
			providerName = "helm_google"
		}
	}
	fullFileName := fmt.Sprintf("%s/%s.tf", testProvidersPath, providerName)
	if _, err := os.Stat(fullFileName); err != nil && errors.Is(err, fs.ErrNotExist) {
		fullFileName = fmt.Sprintf("%s/%s.tf", providersPath, providerName)
		if _, err := os.Stat(fullFileName); err != nil && errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("Provider file not found for %s\n", providerName)
			return []*hclwrite.Block{}
		}
	}
	providerFile := ReadTerraformFile(t, fullFileName)
	return providerFile.Body().Blocks()
}

func addVariables(variables []string, testModuleBody *hclwrite.Body) {
	for _, variable := range variables {
		if variable == "" {
			continue
		}
		testModuleBody.SetAttributeRaw(variable, getTokens("var."+variable))
	}
}

func addOutputs(outputBlocks []*hclwrite.Block, testFileBody *hclwrite.Body) {
	for _, block := range outputBlocks {
		if (len(block.Labels())) == 0 {
			continue
		}
		label := block.Labels()[0]
		if label == "" {
			continue
		}
		attr := block.Body().Attributes()
		output := testFileBody.AppendNewBlock("output", []string{label})
		output.Body().SetAttributeRaw("value", getTokens("module.test."+label))
		addOutputSensitiveAttribute(attr, output.Body())
	}
}

func addOutputSensitiveAttribute(attr map[string]*hclwrite.Attribute, body *hclwrite.Body) {
	for name, _ := range attr {
		if name != "sensitive" {
			continue
		}
		body.SetAttributeValue(name, cty.BoolVal(true))
	}
}

func getTokens(value string) hclwrite.Tokens {
	return getBytesTokens([]byte(value))
}

func getBytesTokens(bytes []byte) hclwrite.Tokens {
	return hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenIdent,
			Bytes: bytes,
		},
	}
}

func ReadTerraformFile(t *testing.T, fileName string) *hclwrite.File {
	file, err := os.ReadFile(fileName)
	if err != nil {
		require.NoError(t, err, "Error reading %s", fileName)
	}
	hclFile, diags := hclwrite.ParseConfig(file, fileName, hcl.InitialPos)
	if diags.HasErrors() {
		require.NoError(t, diags, "Error parsing %s", fileName)
	}
	return hclFile
}

func WriteTerraformFile(t *testing.T, path string, fileName string, bytes []byte) {
	fullFileName := fmt.Sprintf("%s/%s", path, fileName)
	tfFile, err := os.Create(fullFileName)
	require.NoError(t, err, "Error creating %s", fullFileName)
	_, err = tfFile.Write(hclwrite.Format(bytes))
	require.NoError(t, err, "%s file write error", fullFileName)
}
