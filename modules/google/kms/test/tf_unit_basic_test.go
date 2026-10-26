package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/env"
	_ "github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestKms(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		projectID := e.MustProject(t)

		prefix := outputs.String(t, "kms__prefix")
		assert.NotEmpty(t, prefix)

		location := outputs.String(t, "kms__location")
		assert.NotEmpty(t, location)

		keyRingName := outputs.String(t, "kms__key_ring_name")
		assert.NotEmpty(t, keyRingName)

		keyRingId := outputs.String(t, "kms__key_ring_id")
		assert.Contains(t, keyRingId, fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", projectID, location, keyRingName))

		dataKeyId := outputs.String(t, "kms__data_key_id")
		assert.Contains(t, dataKeyId, fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s-data-", projectID, location, keyRingName, prefix))

		configKeyId := outputs.String(t, "kms__config_key_id")
		assert.Contains(t, configKeyId, fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s-config-", projectID, location, keyRingName, prefix))

		telemetryKeyId := outputs.String(t, "kms__telemetry_key_id")
		assert.Contains(t, telemetryKeyId, fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s-telemetry-", projectID, location, keyRingName, prefix))
	})
}
