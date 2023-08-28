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

func TestTerraformEks(t *testing.T) {
	awsRegion = commonAWS.SetupBucket(t, bucketName)
	t.Run("Biz", testTerraformEksBiz)
	t.Run("Pri", testTerraformEksPri)
}

func testTerraformEksBiz(t *testing.T) {
	testTerraformEks(t, "tf_unit_basic_test_biz.tfvars", "biz")
}

func testTerraformEksPri(t *testing.T) {
	testTerraformEks(t, "tf_unit_basic_test_pri.tfvars", "pri")
}

func testTerraformEks(t *testing.T, varFile string, workspaceName string) {
	t.Parallel()
	outputs, destroyFunc := tf.ApplyTerraform(t, bucketName, awsRegion, varFile, workspaceName)
	defer destroyFunc() // Defer needs to be called in outermost function
	clusterName := outputs["cluster_name"]
	assert.Equal(t, fmt.Sprintf("%s-%s", os.Getenv("TF_VAR_prefix"), workspaceName), clusterName,
		"Wrong cluster_name returned")
}
