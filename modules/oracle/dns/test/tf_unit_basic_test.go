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

	zoneId := tf.GetStringValue(t, outputs, "dns__zone_id")
	assert.NotEmpty(t, zoneId, "zone_id was not returned")

	domain := tf.GetStringValue(t, outputs, "dns__domain")
	assert.Equal(t, "biz.biz.internal.test", domain, "Wrong value for domain returned")

	intDomain := tf.GetStringValue(t, outputs, "dns__int_domain")
	assert.Equal(t, domain, intDomain, "int_domain must equal domain (no private-zone split yet)")

	nameServers := tf.GetStringListValue(t, outputs, "dns__name_servers")
	assert.NotEmpty(t, nameServers, "name_servers was not returned")
}
