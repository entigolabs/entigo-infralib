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
	outputs := aws.GetTFOutputs(t, "biz", "infra")
	cluster_name := tf.GetStringValue(t, outputs, "eks__cluster_name")
	assert.Equal(t, fmt.Sprintf("biz-%s-eks",strings.ToLower(os.Getenv("STEP_NAME"))), cluster_name, "Wrong cluster_name returned")
}

func testTerraformEksPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri", "infra")
	cluster_name := tf.GetStringValue(t, outputs, "eks__cluster_name")
	assert.Equal(t, fmt.Sprintf("pri-%s-eks",strings.ToLower(os.Getenv("STEP_NAME"))), cluster_name, "Wrong cluster_name returned")
}
