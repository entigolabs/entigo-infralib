package test

import (
	"testing"
	"fmt"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/stretchr/testify/assert"
)

func TestTerraformEks(t *testing.T) {
	t.Run("Biz", testTerraformEksBiz)
	t.Run("Pri", testTerraformEksPri)
}

func testTerraformEksBiz(t *testing.T) {
	//t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz", "infra")
	cluster_name := aws.GetStringValue(t, outputs, "eks__cluster_name")
	assert.Equal(t, fmt.Sprintf("biz-infra-eks"), cluster_name, "Wrong cluster_name returned")
}

func testTerraformEksPri(t *testing.T) {
	//t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri", "infra")
	cluster_name := aws.GetStringValue(t, outputs, "eks__cluster_name")
	assert.Equal(t, fmt.Sprintf("pri-infra-eks"), cluster_name, "Wrong cluster_name returned")
}
