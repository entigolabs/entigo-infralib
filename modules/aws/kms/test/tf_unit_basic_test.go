package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestKms(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		assert.NotEmpty(t, outputs.String(t, "kms__config_alias_arn"), "config_alias_arn was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__config_key_arn"), "config_key_arn was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__config_key_id"), "config_key_id was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__config_key_policy"), "config_key_policy was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__data_alias_arn"), "data_alias_arn was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__data_key_arn"), "data_key_arn was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__data_key_id"), "data_key_id was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__data_key_policy"), "data_key_policy was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__telemetry_alias_arn"), "telemetry_alias_arn was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__telemetry_key_arn"), "telemetry_key_arn was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__telemetry_key_id"), "telemetry_key_id was not returned")
		assert.NotEmpty(t, outputs.String(t, "kms__telemetry_key_policy"), "telemetry_key_policy was not returned")
	})
}
