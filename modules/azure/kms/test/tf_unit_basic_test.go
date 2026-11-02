package test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestKms(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"azure_biz": testKmsBiz,
	})
}

func testKmsBiz(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	stepName := strings.TrimSpace(strings.ToLower(env.ModulePlacement(t, e).Step.Name))
	prefix := fmt.Sprintf("%s-%s-kms", e.Prefix, stepName)

	// Key vault names are at most 24 characters: the prefix cut to 15 plus the unique suffix.
	short := prefix
	if len(short) > 15 {
		short = short[:15]
	}
	vaultName := fmt.Sprintf("%s-%s", strings.TrimRight(short, "-"), azure.UniqueSuffix(e.Prefix, e.MustSubscriptionID(t), e.MustRegion(t)))
	assert.Equal(t, vaultName, outputs.String(t, "kms__key_vault_name"))
	assert.Equal(t, fmt.Sprintf("https://%s.vault.azure.net/", vaultName), outputs.String(t, "kms__key_vault_uri"))

	versioned := regexp.MustCompile(`/keys/[^/]+/[0-9a-f]{32}$`)
	for _, key := range []string{"data", "config", "telemetry"} {
		keyId := outputs.String(t, fmt.Sprintf("kms__%s_key_id", key))
		assert.Equal(t, fmt.Sprintf("https://%s.vault.azure.net/keys/%s-%s", vaultName, prefix, key), keyId, "%s_key_id must be versionless", key)
		assert.Regexp(t, versioned, outputs.String(t, fmt.Sprintf("kms__%s_key_version_id", key)))
		assert.NotEmpty(t, outputs.String(t, fmt.Sprintf("kms__%s_key_resource_id", key)))
	}
}
