package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"fmt"
)


func TestTerraformDns(t *testing.T) {
	t.Run("Biz", testTerraformDnsBiz)
	t.Run("Pri", testTerraformDnsPri)
}

func testTerraformDnsBiz(t *testing.T) {
	t.Parallel()
        outputs := google.GetTFOutputs(t, "biz")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	
	pubZoneId := tf.GetStringValue(t, outputs, "dns__pub_zone_id")
	assert.Equal(t, fmt.Sprintf("biz-%s-dns-gcp-infralib-entigo-io", stepName), pubZoneId, "Wrong value for pub_zone_id returned")
	
	pubDomain := tf.GetStringValue(t, outputs, "dns__pub_domain")
	assert.Equal(t, fmt.Sprintf("biz-%s-dns.gcp.infralib.entigo.io", stepName), pubDomain, "Wrong value for pub_domain returned")
	
	intZoneId := tf.GetStringValue(t, outputs, "dns__int_zone_id")
	assert.Equal(t, fmt.Sprintf("biz-%s-dns-int-gcp-infralib-entigo-io", stepName), intZoneId, "Wrong value for int_zone_id returned")
	
	intDomain := tf.GetStringValue(t, outputs, "dns__int_domain")
	assert.Equal(t, fmt.Sprintf("biz-%s-dns-int.gcp.infralib.entigo.io", stepName), intDomain, "Wrong value for int_domain returned")
	
	parentZoneId := tf.GetStringValue(t, outputs, "dns__parent_zone_id")
	assert.Equal(t, "gcp-infralib-entigo-io", parentZoneId, "Wrong value for parent_zone_id returned")
	
}

func testTerraformDnsPri(t *testing.T) {
	t.Parallel()
        outputs := google.GetTFOutputs(t, "pri")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	
	pubZoneId := tf.GetStringValue(t, outputs, "dns__pub_zone_id")
	assert.Equal(t, fmt.Sprintf("pri-%s-dns-gcp-infralib-entigo-io", stepName), pubZoneId, "Wrong value for pub_zone_id returned")
	
	pubDomain := tf.GetStringValue(t, outputs, "dns__pub_domain")
	assert.Equal(t, fmt.Sprintf("pri-%s-dns.gcp.infralib.entigo.io", stepName), pubDomain, "Wrong value for pub_domain returned")
	
	intZoneId := tf.GetStringValue(t, outputs, "dns__int_zone_id")
	assert.Equal(t, pubZoneId, intZoneId, "Wrong value for int_zone_id returned")
	
	intDomain := tf.GetStringValue(t, outputs, "dns__int_domain")
	assert.Equal(t, pubDomain, intDomain, "Wrong value for int_domain returned")
	
	parentZoneId := tf.GetStringValue(t, outputs, "dns__parent_zone_id")
	assert.Equal(t, "gcp-infralib-entigo-io", parentZoneId, "Wrong value for parent_zone_id returned")
}


