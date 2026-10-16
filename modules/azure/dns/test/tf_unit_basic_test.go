package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const parentZone = "azure.infralib.entigo.io"

func TestTerraformDns(t *testing.T) {
	t.Run("Biz", testTerraformDnsBiz)
	t.Run("Pri", testTerraformDnsPri)
}

func testTerraformDnsBiz(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "biz")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	pubDomain := fmt.Sprintf("biz-%s-dns.%s", stepName, parentZone)
	intDomain := fmt.Sprintf("biz-%s-dns-int.%s", stepName, parentZone)

	assert.Equal(t, pubDomain, tf.GetStringValue(t, outputs, "dns__pub_domain"))
	assert.Equal(t, intDomain, tf.GetStringValue(t, outputs, "dns__int_domain"))
	assert.Equal(t, intDomain, tf.GetStringValue(t, outputs, "dns__int_private_domain"))
	assert.Equal(t, intDomain, tf.GetStringValue(t, outputs, "dns__int_cert_zone_name"))
	pubZoneId := tf.GetStringValue(t, outputs, "dns__pub_zone_id")
	intZoneId := tf.GetStringValue(t, outputs, "dns__int_zone_id")
	assert.NotEqual(t, pubZoneId, intZoneId, "int_zone_id and pub_zone_id must not be equal")
	assert.Contains(t, strings.ToLower(pubZoneId), "/dnszones/")
	assert.Contains(t, strings.ToLower(intZoneId), "/privatednszones/")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "dns__int_cert_zone_id"), "int_cert_zone_id was not returned")
	assert.Equal(t, 3, len(tf.GetValue(t, outputs, "dns__zone_ids").(map[string]interface{})), "Wrong number of zone_ids")
	assert.Equal(t, 1, len(tf.GetValue(t, outputs, "dns__validation_zone_ids").(map[string]interface{})), "Wrong number of validation_zone_ids")

	parentRg := fmt.Sprintf("infralib-permanent-%s", azure.Location())
	require.NoError(t, azure.WaitUntilDnsRecordExists(t, parentRg, parentZone, fmt.Sprintf("biz-%s-dns", stepName), "NS", 10, 6*time.Second))
	require.NoError(t, azure.WaitUntilDnsRecordExists(t, parentRg, parentZone, fmt.Sprintf("biz-%s-dns-int", stepName), "NS", 10, 6*time.Second))
}

func testTerraformDnsPri(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "pri")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	domain := fmt.Sprintf("pri-%s-dns.%s", stepName, parentZone)

	assert.Equal(t, domain, tf.GetStringValue(t, outputs, "dns__pub_domain"))
	assert.Equal(t, domain, tf.GetStringValue(t, outputs, "dns__int_domain"))
	assert.Equal(t, "", tf.GetStringValue(t, outputs, "dns__int_private_domain"))
	assert.Equal(t, "", tf.GetStringValue(t, outputs, "dns__int_cert_zone_id"))
	assert.Equal(t, tf.GetStringValue(t, outputs, "dns__pub_zone_id"), tf.GetStringValue(t, outputs, "dns__int_zone_id"), "int_zone_id and pub_zone_id must be equal")

	parentRg := fmt.Sprintf("infralib-permanent-%s", azure.Location())
	require.NoError(t, azure.WaitUntilDnsRecordExists(t, parentRg, parentZone, fmt.Sprintf("pri-%s-dns", stepName), "NS", 10, 6*time.Second))
}
