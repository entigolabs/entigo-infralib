package test

import (
	"fmt"
	"testing"

	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	"github.com/gruntwork-io/terratest/modules/logger"
)

const bucketName = "infralib-modules-gcp-dns-tf"

var Region string

func TestTerraformDns(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformDnsBiz)
	t.Run("Pri", testTerraformDnsPri)
}

func testTerraformDnsBiz(t *testing.T) {
	ctxx := context.Background()
	creds, err := google.FindDefaultCredentials(ctxx)
	if err != nil {
		logger.Logf(t, "Failed to find default credentials: %v", err)
	}
	crmService, err := cloudresourcemanager.NewService(ctxx, option.WithCredentials(creds))
	if err != nil {
		logger.Logf(t, "Failed to create cloudresourcemanager service: %v", err)
	}
	projectID := "entigo-infralib2"
	if projectID == "" {
		// If ProjectID is empty, fetch the list of projects
		projectListCall := crmService.Projects.List()
		projectList, err := projectListCall.Do()
		if err != nil {
			logger.Logf(t, "Failed to list projects: %v", err)
		}
		if len(projectList.Projects) > 0 {
			projectID = projectList.Projects[0].ProjectId
		}
	}
        
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
	network := fmt.Sprintf("%s",result.Payload.Data)

	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
	        "vpc_id":                network,
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
