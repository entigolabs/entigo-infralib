package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/stretchr/testify/assert"
)


func TestTerraformDns(t *testing.T) {
	t.Run("Biz", testTerraformDnsBiz)
	t.Run("Pri", testTerraformDnsPri)
}

func testTerraformDnsBiz(t *testing.T) {
	t.Parallel()
        outputs := google.GetTFOutputs(t, "biz", "net")
	
	pubZoneId := tf.GetStringValue(t, outputs, "dns__pub_zone_id")
	assert.Equal(t, "biz-net-dns-gcp-infralib-entigo-io", pubZoneId, "Wrong value for pub_zone_id returned")
	
	pubDomain := tf.GetStringValue(t, outputs, "dns__pub_domain")
	assert.Equal(t, "biz-net-dns.gcp.infralib.entigo.io", pubDomain, "Wrong value for pub_domain returned")
	
	intZoneId := tf.GetStringValue(t, outputs, "dns__int_zone_id")
	assert.Equal(t, "biz-net-dns-int-gcp-infralib-entigo-io", intZoneId, "Wrong value for int_zone_id returned")
	
	intDomain := tf.GetStringValue(t, outputs, "dns__int_domain")
	assert.Equal(t, "biz-net-dns-int.gcp.infralib.entigo.io", intDomain, "Wrong value for int_domain returned")
	
	parentZoneId := tf.GetStringValue(t, outputs, "dns__parent_zone_id")
	assert.Equal(t, "gcp-infralib-entigo-io", parentZoneId, "Wrong value for parent_zone_id returned")
	
}

func testTerraformDnsPri(t *testing.T) {
	t.Parallel()
        outputs := google.GetTFOutputs(t, "pri", "net")
	pubZoneId := tf.GetStringValue(t, outputs, "dns__pub_zone_id")
	assert.Equal(t, "pri-net-dns-gcp-infralib-entigo-io", pubZoneId, "Wrong value for pub_zone_id returned")
	
	pubDomain := tf.GetStringValue(t, outputs, "dns__pub_domain")
	assert.Equal(t, "pri-net-dns.gcp.infralib.entigo.io", pubDomain, "Wrong value for pub_domain returned")
	
	intZoneId := tf.GetStringValue(t, outputs, "dns__int_zone_id")
	assert.Equal(t, pubZoneId, intZoneId, "Wrong value for int_zone_id returned")
	
	intDomain := tf.GetStringValue(t, outputs, "dns__int_domain")
	assert.Equal(t, pubDomain, intDomain, "Wrong value for int_domain returned")
	
	parentZoneId := tf.GetStringValue(t, outputs, "dns__parent_zone_id")
	assert.Equal(t, "gcp-infralib-entigo-io", parentZoneId, "Wrong value for parent_zone_id returned")
}


