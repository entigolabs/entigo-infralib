package test

import (
	"fmt"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

const bucketName = "infralib-modules-gcp-dns-tf"

var Region string

func TestTerraformDns(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformDnsBiz)
	t.Run("Pri", testTerraformDnsPri)
}

func testTerraformDnsBiz(t *testing.T) {
	projectID := commonGCP.GetProjectID
	fmt.Printf("Project id is: %s \n", projectID)

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		logger.Logf(t, "failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-vpc_id/versions/latest", projectID),
	}
	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Logf(t, "failed to access secret %v", err)
	}
	fmt.Printf("retrieved payload for: %s %s\n", result.Name, result.Payload.Data)
	network := fmt.Sprintf("%s", result.Payload.Data)

	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"vpc_id": network,
	})
	testTerraformDnsBizAssert(t, "biz", options)
}

func testTerraformDnsPri(t *testing.T) {
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{})
	testTerraformDnsBizAssert(t, "pri", options)
}

func testTerraformDnsBizAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()
	assert.NotEmpty(t, outputs["pub_zone_id"], "pub_zone_id was not returned")
}

func testTerraformDnsPriAssert(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc()
	assert.NotEmpty(t, outputs["pub_domain"], "pub_domain was not returned")
}
