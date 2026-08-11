package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformDns(t *testing.T) {
	t.Run("Biz", testTerraformDnsBiz)
}

func testTerraformDnsBiz(t *testing.T) {
	t.Parallel()
	outputs := oracle.GetTFOutputs(t, "biz")

	zoneId := tf.GetStringValue(t, outputs, "dns__pub_zone_id")
	assert.NotEmpty(t, zoneId, "pub_zone_id was not returned")

	domain := tf.GetStringValue(t, outputs, "dns__pub_domain")
	assert.Equal(t, "biz.biz.internal.test", domain, "Wrong value for pub_domain returned")

	// biz.yaml has no private domain, so int_* falls back to the public default. This is a
	// deliberate deviation from aws-v2/route53, which would reject the configuration - every
	// Oracle app reads .toutput.<dns>.int_domain, so the fallback has to hold.
	assert.Equal(t, domain, tf.GetStringValue(t, outputs, "dns__int_domain"), "int_domain should fall back to pub_domain when no domain is private")
	assert.Equal(t, zoneId, tf.GetStringValue(t, outputs, "dns__int_zone_id"), "int_zone_id should fall back to pub_zone_id when no domain is private")

	// The map outputs carry every domain, not just the defaults.
	zoneIds, ok := tf.GetValue(t, outputs, "dns__zone_ids").(map[string]interface{})
	assert.True(t, ok, "zone_ids was not a map")
	assert.Len(t, zoneIds, 2, "zone_ids should carry both domains")
	assert.NotEmpty(t, zoneIds["secondary"], "the secondary zone was not created")

	domainNames, ok := tf.GetValue(t, outputs, "dns__domain_names").(map[string]interface{})
	assert.True(t, ok, "domain_names was not a map")
	assert.Equal(t, "biz2.biz.internal.test", domainNames["secondary"], "Wrong value for the secondary domain name")

	nameservers, ok := tf.GetValue(t, outputs, "dns__nameservers").(map[string]interface{})
	assert.True(t, ok, "nameservers was not a map")
	assert.Len(t, nameservers, 2, "nameservers should carry both created zones")

	// No certificates: biz.yaml sets create_certificate = false on both domains, and no CA is
	// wired in for a single-module test.
	assert.Empty(t, tf.GetStringValue(t, outputs, "dns__pub_cert_ocid"), "pub_cert_ocid should be empty when no certificate is created")
	assert.Empty(t, tf.GetStringValue(t, outputs, "dns__int_cert_ocid"), "int_cert_ocid should be empty when no certificate is created")
	assert.Empty(t, tf.GetStringValue(t, outputs, "dns__certificate_authority_id"), "certificate_authority_id should be empty when none is wired in")
}
