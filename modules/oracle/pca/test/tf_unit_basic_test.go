package test

import (
	"strings"
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

	// biz.yaml adopts the existing CA by name rather than creating one.
	caId := tf.GetStringValue(t, outputs, "pca__certificate_authority_id")
	assert.True(t, strings.HasPrefix(caId, "ocid1.certificateauthority."), "certificate_authority_id is not a CA OCID: %q", caId)

	caName := tf.GetStringValue(t, outputs, "pca__certificate_authority_name")
	assert.Equal(t, "biz-net-pca-root-ca-3pbbwkcu", caName, "Wrong value for certificate_authority_name returned")

	bundleCommand := tf.GetStringValue(t, outputs, "pca__certificate_authority_bundle_command")
	assert.Contains(t, bundleCommand, caId, "certificate_authority_bundle_command does not name the adopted CA")
}
