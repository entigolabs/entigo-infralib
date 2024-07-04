package test

import (
	"fmt"
	"os"
	"testing"

	commonGCP "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-gce-gke-tf"

var Region string

func TestTerraformGke(t *testing.T) {
	Region = commonGCP.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformGkeBiz)
	t.Run("Pri", testTerraformGkePri)
}

func testTerraformGkeBiz(t *testing.T) {
	projectID := os.Getenv("GOOGLE_PROJECT")
	fmt.Printf("Project id is: %s \n", projectID)

	network := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-vpc_name/versions/latest", projectID))
	subnetwork := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-private_subnets/versions/latest", projectID))
	subnetworkPods := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-private_subnets_pods/versions/latest", projectID))
	subnetworkServices := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-biz-private_subnets_services/versions/latest", projectID))

	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"network":           network,
		"subnetwork":        subnetwork,
		"ip_range_pods":     subnetworkPods,
		"ip_range_services": subnetworkServices,
	})
	testTerraformGke(t, "biz", options)
}

func testTerraformGkePri(t *testing.T) {
	projectID := os.Getenv("GOOGLE_PROJECT")
	fmt.Printf("Project id is: %s \n", projectID)

	network := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-pri-vpc_name/versions/latest", projectID))
	subnetwork := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-pri-private_subnets/versions/latest", projectID))
	subnetworkPods := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-pri-private_subnets_pods/versions/latest", projectID))
	subnetworkServices := commonGCP.GetSecret(t, fmt.Sprintf("projects/%s/secrets/entigo-infralib-runner-main-pri-private_subnets_services/versions/latest", projectID))

	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"network":           network,
		"subnetwork":        subnetwork,
		"ip_range_pods":     subnetworkPods,
		"ip_range_services": subnetworkServices,
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
