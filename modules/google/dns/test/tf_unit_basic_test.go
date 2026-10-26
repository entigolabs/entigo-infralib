package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/env"
	_ "github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestDns(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		modulePrefix := fmt.Sprintf("%s-%s-dns", e.Prefix, env.ModulePlacement(t, e).Step.Name)

		pubZoneId := outputs.String(t, "dns__pub_zone_id")
		assert.Equal(t, modulePrefix+"-gcp-infralib-entigo-io", pubZoneId, "Wrong value for pub_zone_id returned")

		pubDomain := outputs.String(t, "dns__pub_domain")
		assert.Equal(t, modulePrefix+".gcp.infralib.entigo.io", pubDomain, "Wrong value for pub_domain returned")

		intZoneId := outputs.String(t, "dns__int_zone_id")
		intDomain := outputs.String(t, "dns__int_domain")
		if e.Prefix == "biz" {
			// biz creates a separate internal zone.
			assert.Equal(t, modulePrefix+"-int-gcp-infralib-entigo-io", intZoneId, "Wrong value for int_zone_id returned")
			assert.Equal(t, modulePrefix+"-int.gcp.infralib.entigo.io", intDomain, "Wrong value for int_domain returned")
		} else {
			// pri has no internal zone, so the public one serves as both.
			assert.Equal(t, pubZoneId, intZoneId, "Wrong value for int_zone_id returned")
			assert.Equal(t, pubDomain, intDomain, "Wrong value for int_domain returned")
		}

		parentZoneId := outputs.String(t, "dns__parent_zone_id")
		assert.Equal(t, "gcp-infralib-entigo-io", parentZoneId, "Wrong value for parent_zone_id returned")
	})
}
