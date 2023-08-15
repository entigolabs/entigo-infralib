package test

import (
	"fmt"
	commonAWS "github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const bucketName = "infralib-modules-aws-eks-tf"

var awsRegion string

func TestEKSRunner(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformBasicBiz)
	t.Run("Pri", testTerraformBasicPri)
}

func testTerraformBasicBiz(t *testing.T) {
	testTerraformBasic(t, "tf_unit_basic_test_biz.tfvars", "biz")
}

func testTerraformBasicPri(t *testing.T) {
	testTerraformBasic(t, "tf_unit_basic_test_pri.tfvars", "pri")
}

func testTerraformBasic(t *testing.T, varFile string, workspaceName string) {
	t.Parallel()
	outputs := tf.ApplyTerraform(t, bucketName, awsRegion, varFile, workspaceName)
	clusterName := outputs["cluster_name"]
	assert.Equal(t, fmt.Sprintf("%s-%s", os.Getenv("TF_VAR_prefix"), workspaceName), clusterName,
		"Wrong cluster_name returned")
}
