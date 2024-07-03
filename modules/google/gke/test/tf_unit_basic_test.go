package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

const bucketName = "infralib-modules-gce-gke-tf"

var Region string

func TestTerraformGke(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformGkeBiz)
	t.Run("Pri", testTerraformGkePri)
}

func testTerraformGkeBiz(t *testing.T) {
	ctxx := context.Background()
	creds, err := google.FindDefaultCredentials(ctxx)
	if err != nil {
		logger.Logf(t, "Failed to find default credentials: %v", err)
	}
	crmService, err := cloudresourcemanager.NewService(ctxx, option.WithCredentials(creds))
	if err != nil {
		logger.Logf(t, "Failed to create cloudresourcemanager service: %v", err)
	}
	projectID := creds.ProjectID
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
		Name: fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-vpc_name/versions/latest", projectID),
	}
	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Logf(t, "failed to access secret %v", err)
	}
	fmt.Printf("retrieved payload for: %s %s\n", result.Name, result.Payload.Data)
	network := fmt.Sprintf("%s", result.Payload.Data)

	// Build the request.
	req = &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-private_subnets/versions/latest", projectID),
	}
	// Call the API.
	result, err = client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Logf(t, "failed to access secret %v", err)
	}
	fmt.Printf("retrieved payload for: %s %s\n", result.Name, result.Payload.Data)
	subnetwork := strings.Trim(strings.Split(fmt.Sprintf("%s", result.Payload.Data), ",")[0], `"`)

	// Build the request.
	req = &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-private_subnets_pods/versions/latest", projectID),
	}
	// Call the API.
	result, err = client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Logf(t, "failed to access secret %v", err)
	}
	fmt.Printf("retrieved payload for: %s %s\n", result.Name, result.Payload.Data)
	subnetworkpods := strings.Trim(strings.Split(fmt.Sprintf("%s", result.Payload.Data), ",")[0], `"`)

	// Build the request.
	req = &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-private_subnets_services/versions/latest", projectID),
	}
	// Call the API.
	result, err = client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Logf(t, "failed to access secret %v", err)
	}
	fmt.Printf("retrieved payload for: %s %s\n", result.Name, result.Payload.Data)
	subnetworkservices := strings.Trim(strings.Split(fmt.Sprintf("%s", result.Payload.Data), ",")[0], `"`)

	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"network":           network,
		"subnetwork":        subnetwork,
		"ip_range_pods":     subnetworkpods,
		"ip_range_services": subnetworkservices,
	})
	testTerraformGke(t, "biz", options)
}

func testTerraformGkePri(t *testing.T) {
	ctxx := context.Background()
	creds, err := google.FindDefaultCredentials(ctxx)
	if err != nil {
		logger.Logf(t, "Failed to find default credentials: %v", err)
	}
	crmService, err := cloudresourcemanager.NewService(ctxx, option.WithCredentials(creds))
	if err != nil {
		logger.Logf(t, "Failed to create cloudresourcemanager service: %v", err)
	}
	projectID := creds.ProjectID
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
		Name: fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-pri-vpc_id/versions/latest", projectID),
	}
	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Logf(t, "failed to access secret %v", err)
	}
	fmt.Printf("retrieved payload for: %s %s\n", result.Name, result.Payload.Data)
	network := fmt.Sprintf("%s", result.Payload.Data)
	network = network[strings.LastIndex(network, "/")+1:]

	// Build the request.
	req = &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-pri-private_subnets/versions/latest", projectID),
	}
	// Call the API.
	result, err = client.AccessSecretVersion(ctx, req)
	if err != nil {
		logger.Logf(t, "failed to access secret %v", err)
	}
	fmt.Printf("retrieved payload for: %s %s\n", result.Name, result.Payload.Data)
	subnetwork := strings.Trim(strings.Split(fmt.Sprintf("%s", result.Payload.Data), ",")[0], `"`)
	subnetwork = subnetwork[strings.LastIndex(subnetwork, "/")+1:]

	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"network":           network,
		"subnetwork":        subnetwork,
		"ip_range_pods":     fmt.Sprintf("%s-pods", network),
		"ip_range_services": fmt.Sprintf("%s-services", network),
	})
	testTerraformGke(t, "pri", options)
}

func testTerraformGke(t *testing.T, workspaceName string, options *terraform.Options) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, workspaceName, options)
	defer destroyFunc() // Defer needs to be called in outermost function
	clusterName := outputs["cluster_name"]
	assert.Equal(t, fmt.Sprintf("%s-%s", os.Getenv("TF_VAR_prefix"), workspaceName), clusterName,
		"Wrong cluster_name returned")
}
