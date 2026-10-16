package test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformKms(t *testing.T) {
	t.Run("Biz", testTerraformKmsBiz)
}

func testTerraformKmsBiz(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "biz")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	prefix := fmt.Sprintf("biz-%s-kms", stepName)

	short := prefix
	if len(short) > 15 {
		short = short[:15]
	}
	vaultName := fmt.Sprintf("%s-%s", strings.TrimRight(short, "-"), azure.UniqueSuffix("biz", azure.SubscriptionID(), azure.Location()))
	assert.Equal(t, vaultName, tf.GetStringValue(t, outputs, "kms__key_vault_name"))
	assert.Equal(t, fmt.Sprintf("https://%s.vault.azure.net/", vaultName), tf.GetStringValue(t, outputs, "kms__key_vault_uri"))

	versioned := regexp.MustCompile(`/keys/[^/]+/[0-9a-f]{32}$`)
	for _, key := range []string{"data", "config", "telemetry"} {
		keyId := tf.GetStringValue(t, outputs, fmt.Sprintf("kms__%s_key_id", key))
		assert.Equal(t, fmt.Sprintf("https://%s.vault.azure.net/keys/%s-%s", vaultName, prefix, key), keyId, "%s_key_id must be versionless", key)
		assert.Regexp(t, versioned, tf.GetStringValue(t, outputs, fmt.Sprintf("kms__%s_key_version_id", key)))
		assert.NotEmpty(t, tf.GetStringValue(t, outputs, fmt.Sprintf("kms__%s_key_resource_id", key)))
	}
}
