package test

import (
	"fmt"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformKms(t *testing.T) {
	t.Run("Biz", testTerraformKmsBiz)
}

func testTerraformKmsBiz(t *testing.T) {
	t.Parallel()
	outputs := google.GetTFOutputs(t, "biz")

	prefix := tf.GetStringValue(t, outputs, "kms__prefix")
	assert.NotEmpty(t, prefix)

	location := tf.GetStringValue(t, outputs, "kms__location")
	assert.NotEmpty(t, location)

	keyRingName := tf.GetStringValue(t, outputs, "kms__key_ring_name")
	assert.NotEmpty(t, keyRingName)

	keyRingId := tf.GetStringValue(t, outputs, "kms__key_ring_id")
	assert.Contains(t, keyRingId, fmt.Sprintf("/locations/%s/keyRings/%s", location, keyRingName))

	dataKeyId := tf.GetStringValue(t, outputs, "kms__data_key_id")
	assert.Contains(t, dataKeyId, fmt.Sprintf("/locations/%s/keyRings/%s/cryptoKeys/%s-data-", location, keyRingName, prefix))

	configKeyId := tf.GetStringValue(t, outputs, "kms__config_key_id")
	assert.Contains(t, configKeyId, fmt.Sprintf("/locations/%s/keyRings/%s/cryptoKeys/%s-config-", location, keyRingName, prefix))

	telemetryKeyId := tf.GetStringValue(t, outputs, "kms__telemetry_key_id")
	assert.Contains(t, telemetryKeyId, fmt.Sprintf("/locations/%s/keyRings/%s/cryptoKeys/%s-telemetry-", location, keyRingName, prefix))
}
