package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/stretchr/testify/assert"
)

func TestTerraformKms(t *testing.T) {

	t.Run("Biz", testTerraformKmsBiz)
}

func testTerraformKmsBiz(t *testing.T) {
        t.Parallel()
        outputs := aws.GetTFOutputs(t, "biz", "net")
	
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__config_alias_arn"), "config_alias_arn was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__config_key_arn"), "config_key_arn was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__config_key_id"), "config_key_id was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__config_key_policy"), "config_key_policy was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__data_alias_arn"), "data_alias_arn was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__data_key_arn"), "data_key_arn was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__data_key_id"), "data_key_id was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__data_key_policy"), "data_key_policy was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__telemetry_alias_arn"), "telemetry_alias_arn was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__telemetry_key_arn"), "telemetry_key_arn was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__telemetry_key_id"), "telemetry_key_id was not returned")
	assert.NotEmpty(t, aws.GetStringValue(t, outputs, "kms__telemetry_key_policy"), "telemetry_key_policy was not returned")
      
}

