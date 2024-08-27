package test

import (
	"fmt"
	"os"
	"testing"

	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
)

const bucketName = "infralib-modules-aws-eks-desired-size-tf"

var awsRegion string

func TestTerraformEksDesiredSize(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformEksDesiredSizeBiz)
	t.Run("Pri", testTerraformEksDesiredSizePri)
}

func testTerraformEksDesiredSizeBiz(t *testing.T) {
	t.Parallel()
	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{
		"cluster_name": "runner-main-biz",
	})
	_, destroyFunc := tf.ApplyTerraform(t, "biz", options)
	defer destroyFunc()

	prefix := os.Getenv("TF_VAR_prefix")
	workspaceName := "biz"

	eks_main_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_main_min_size", prefix, workspaceName))
	eks_mainarm_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_mainarm_min_size", prefix, workspaceName))
	eks_tools_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_tools_min_size", prefix, workspaceName))
	eks_mon_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_mon_min_size", prefix, workspaceName))
	eks_db_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_db_min_size", prefix, workspaceName))
	eks_spot_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_spot_min_size", prefix, workspaceName))
	eks_altarm_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_altarm_min_size", prefix, workspaceName))

	assert.Equal(t, "4", eks_main_min_size, "Wrong value for eks_main_min_size returned")
	assert.Equal(t, "1", eks_mainarm_min_size, "Wrong value for eks_mainarm_min_size returned")
	assert.Equal(t, "2", eks_tools_min_size, "Wrong value for eks_tools_min_size returned")
	assert.Equal(t, "1", eks_mon_min_size, "Wrong value for eks_mon_min_size returned")
	assert.Equal(t, "0", eks_db_min_size, "Wrong value for eks_db_min_size returned")
	assert.Equal(t, "0", eks_spot_min_size, "Wrong value for eks_spot_min_size returned")
	assert.Equal(t, "1", eks_altarm_min_size, "Wrong value for eks_altarm_min_size returned")
}

func testTerraformEksDesiredSizePri(t *testing.T) {
	t.Parallel()

	prefix := os.Getenv("TF_VAR_prefix")
	workspaceName := "pri"

	options := tf.InitAWSTerraform(t, bucketName, awsRegion, "tf_unit_basic_test_pri.tfvars", map[string]interface{}{
		"cluster_name": "runner-main-pri",
	})
	_, destroyFunc := tf.ApplyTerraform(t, "pri", options)
	defer destroyFunc()

	eks_main_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_main_min_size", prefix, workspaceName))
	eks_mainarm_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_mainarm_min_size", prefix, workspaceName))
	eks_tools_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_tools_min_size", prefix, workspaceName))
	eks_mon_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_mon_min_size", prefix, workspaceName))
	eks_db_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_db_min_size", prefix, workspaceName))
	eks_spot_min_size := aws.GetParameter(t, awsRegion, fmt.Sprintf("/entigo-infralib/%s-%s/eks_spot_min_size", prefix, workspaceName))

	assert.Equal(t, "4", eks_main_min_size, "Wrong value for eks_main_min_size returned")
	assert.Equal(t, "0", eks_mainarm_min_size, "Wrong value for eks_mainarm_min_size returned")
	assert.Equal(t, "0", eks_tools_min_size, "Wrong value for eks_tools_min_size returned")
	assert.Equal(t, "0", eks_mon_min_size, "Wrong value for eks_mon_min_size returned")
	assert.Equal(t, "0", eks_db_min_size, "Wrong value for eks_db_min_size returned")
	assert.Equal(t, "0", eks_spot_min_size, "Wrong value for eks_spot_min_size returned")
}
