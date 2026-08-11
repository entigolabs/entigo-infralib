package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformPca(t *testing.T) {
	t.Run("Biz", testTerraformPcaBiz)
}

func testTerraformPcaBiz(t *testing.T) {
	t.Parallel()
	outputs := oracle.GetTFOutputs(t, "biz")

	// biz.yaml turns the CA off - see the comment there for why it is not exercised here.
	caId := tf.GetStringValue(t, outputs, "pca__certificate_authority_id")
	assert.Empty(t, caId, "certificate_authority_id should be empty when create_ca is false")

	caName := tf.GetStringValue(t, outputs, "pca__certificate_authority_name")
	assert.Empty(t, caName, "certificate_authority_name should be empty when create_ca is false")

	bundleCommand := tf.GetStringValue(t, outputs, "pca__certificate_authority_bundle_command")
	assert.Empty(t, bundleCommand, "certificate_authority_bundle_command should be empty when create_ca is false")
}
