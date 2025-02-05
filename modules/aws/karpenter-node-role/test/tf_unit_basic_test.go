package test

import (
	"testing"
	"fmt"
	"os"
	"strings"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformEks(t *testing.T) {
	t.Run("Biz", testTerraformEksBiz)
	t.Run("Pri", testTerraformEksPri)
}

func testTerraformEksBiz(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz")
	role_name := tf.GetStringValue(t, outputs, "karpenter-node-role__role_name")
	assert.Equal(t, fmt.Sprintf("biz-%s-karpenter-node-role",strings.ToLower(os.Getenv("STEP_NAME"))), role_name, "Wrong role_name returned")
	role_arn := tf.GetStringValue(t, outputs, "karpenter-node-role__role_arn")
	assert.NotEmpty(t, role_arn, "Empty role_arn returned")
}

func testTerraformEksPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri")
	role_name := tf.GetStringValue(t, outputs, "karpenter-node-role__role_name")
	assert.Equal(t, fmt.Sprintf("pri-%s-karpenter-node-role",strings.ToLower(os.Getenv("STEP_NAME"))), role_name, "Wrong role_name returned")
	role_arn := tf.GetStringValue(t, outputs, "karpenter-node-role__role_arn")
	assert.NotEmpty(t, role_arn, "Empty role_arn returned")
}
