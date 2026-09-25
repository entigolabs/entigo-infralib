package test

import (
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformKms(t *testing.T) {
	t.Run("Biz", testTerraformKmsBiz)
}

func testTerraformKmsBiz(t *testing.T) {
	t.Parallel()
	outputs := oracle.GetTFOutputs(t, "biz")

	vaultId := tf.GetStringValue(t, outputs, "kms__vault_id")
	assert.NotEmpty(t, vaultId, "vault_id was not returned")

	managementEndpoint := tf.GetStringValue(t, outputs, "kms__vault_management_endpoint")
	assert.NotEmpty(t, managementEndpoint, "vault_management_endpoint was not returned")

	cryptoEndpoint := tf.GetStringValue(t, outputs, "kms__vault_crypto_endpoint")
	assert.NotEmpty(t, cryptoEndpoint, "vault_crypto_endpoint was not returned")

	// The three keys every cloud's kms module provides.
	for _, name := range []string{"data", "config", "telemetry"} {
		keyId := tf.GetStringValue(t, outputs, "kms__"+name+"_key_id")
		assert.NotEmpty(t, keyId, name+"_key_id was not returned")
		assert.True(t, strings.HasPrefix(keyId, "ocid1.key."), "Wrong value for "+name+"_key_id returned")
	}

	// biz.yaml turns the CA key off - see the comment there for why it is not exercised here.
	caKeyId := tf.GetStringValue(t, outputs, "kms__ca_key_id")
	assert.Empty(t, caKeyId, "ca_key_id should be empty when create_ca_key is false")
}
