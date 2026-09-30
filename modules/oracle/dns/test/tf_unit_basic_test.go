package test

import (
	"strings"
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
	assert.Equal(t, "biz-net-dns.oci.infralib.entigo.io", domain, "Wrong value for pub_domain returned")

	// biz.yaml has one public and one private domain, so int_* is the private one.
	intDomain := tf.GetStringValue(t, outputs, "dns__int_domain")
	assert.Equal(t, "biz-net-dns-int.oci.infralib.entigo.io", intDomain, "Wrong value for int_domain returned")
	intZoneId := tf.GetStringValue(t, outputs, "dns__int_zone_id")
	assert.NotEmpty(t, intZoneId, "int_zone_id was not returned")
	assert.NotEqual(t, zoneId, intZoneId, "int_zone_id should be the private zone, not the public one")

	// The map outputs carry every domain, not just the defaults.
	zoneIds, ok := tf.GetValue(t, outputs, "dns__zone_ids").(map[string]interface{})
	assert.True(t, ok, "zone_ids was not a map")
	assert.Len(t, zoneIds, 2, "zone_ids should carry both domains")

	domainNames, ok := tf.GetValue(t, outputs, "dns__domain_names").(map[string]interface{})
	assert.True(t, ok, "domain_names was not a map")
	assert.Equal(t, domain, domainNames["public"], "Wrong value for the public domain name")
	assert.Equal(t, intDomain, domainNames["private"], "Wrong value for the private domain name")

	nameservers, ok := tf.GetValue(t, outputs, "dns__nameservers").(map[string]interface{})
	assert.True(t, ok, "nameservers was not a map")
	assert.Len(t, nameservers, 2, "nameservers should carry both created zones")

	// Both domains get a certificate from the CA modules/oracle/pca adopts.
	assert.True(t, strings.HasPrefix(tf.GetStringValue(t, outputs, "dns__pub_cert_ocid"), "ocid1.certificate."), "pub_cert_ocid is not a certificate OCID")
	assert.True(t, strings.HasPrefix(tf.GetStringValue(t, outputs, "dns__int_cert_ocid"), "ocid1.certificate."), "int_cert_ocid is not a certificate OCID")
	assert.True(t, strings.HasPrefix(tf.GetStringValue(t, outputs, "dns__certificate_authority_id"), "ocid1.certificateauthority."), "certificate_authority_id is not a CA OCID")
}
