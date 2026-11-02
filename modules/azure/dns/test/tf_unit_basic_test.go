package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

// The test zones are delegated from this permanent parent zone.
const (
	parentZone = "azure.infralib.entigo.io"
	parentRg   = "infralib-permanent-westeurope"
)

func TestDns(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"azure_biz": testDnsBiz,
		"azure_pri": testDnsPri,
	})
}

// biz has a separate internal zone with its own validation zone.
func testDnsBiz(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	stepName := strings.TrimSpace(strings.ToLower(env.ModulePlacement(t, e).Step.Name))
	pubName := fmt.Sprintf("%s-%s-dns", e.Prefix, stepName)
	intName := fmt.Sprintf("%s-%s-dns-int", e.Prefix, stepName)
	pubDomain := fmt.Sprintf("%s.%s", pubName, parentZone)
	intDomain := fmt.Sprintf("%s.%s", intName, parentZone)

	assert.Equal(t, pubDomain, outputs.String(t, "dns__pub_domain"))
	assert.Equal(t, intDomain, outputs.String(t, "dns__int_domain"))
	assert.Equal(t, intDomain, outputs.String(t, "dns__int_private_domain"))
	assert.Equal(t, intDomain, outputs.String(t, "dns__int_cert_zone_name"))
	pubZoneId := outputs.String(t, "dns__pub_zone_id")
	intZoneId := outputs.String(t, "dns__int_zone_id")
	assert.NotEqual(t, pubZoneId, intZoneId, "int_zone_id and pub_zone_id must not be equal")
	assert.Contains(t, strings.ToLower(pubZoneId), "/dnszones/")
	assert.Contains(t, strings.ToLower(intZoneId), "/privatednszones/")
	assert.NotEmpty(t, outputs.String(t, "dns__int_cert_zone_id"), "int_cert_zone_id was not returned")
	assert.Equal(t, 3, len(outputs.Map(t, "dns__zone_ids")), "Wrong number of zone_ids")
	assert.Equal(t, 1, len(outputs.Map(t, "dns__validation_zone_ids")), "Wrong number of validation_zone_ids")
	assert.True(t, strings.HasSuffix(strings.ToLower(outputs.String(t, "dns__agc_private_dns_zone_id")), "/privatednszones/privatelink.alb.azure.com"), "agc_private_dns_zone_id must be the privatelink.alb.azure.com zone")

	require.NoError(t, azure.WaitUntilDNSRecordExists(t, e, parentRg, parentZone, pubName, "NS", 10, 6*time.Second))
	require.NoError(t, azure.WaitUntilDNSRecordExists(t, e, parentRg, parentZone, intName, "NS", 10, 6*time.Second))
}

// pri uses the public zone for internal names as well.
func testDnsPri(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	stepName := strings.TrimSpace(strings.ToLower(env.ModulePlacement(t, e).Step.Name))
	name := fmt.Sprintf("%s-%s-dns", e.Prefix, stepName)
	domain := fmt.Sprintf("%s.%s", name, parentZone)

	assert.Equal(t, domain, outputs.String(t, "dns__pub_domain"))
	assert.Equal(t, domain, outputs.String(t, "dns__int_domain"))
	assert.Equal(t, "", outputs.String(t, "dns__int_private_domain"))
	assert.Equal(t, "", outputs.String(t, "dns__int_cert_zone_id"))
	assert.Equal(t, outputs.String(t, "dns__pub_zone_id"), outputs.String(t, "dns__int_zone_id"), "int_zone_id and pub_zone_id must be equal")
	assert.True(t, strings.HasSuffix(strings.ToLower(outputs.String(t, "dns__agc_private_dns_zone_id")), "/privatednszones/privatelink.alb.azure.com"), "agc_private_dns_zone_id must be the privatelink.alb.azure.com zone")

	require.NoError(t, azure.WaitUntilDNSRecordExists(t, e, parentRg, parentZone, name, "NS", 10, 6*time.Second))
}
