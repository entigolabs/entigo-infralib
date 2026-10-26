package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestAgc(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		netOutputs := tf.GetStep(t, e, "net")

		prefix := fmt.Sprintf("%s-%s-agc", e.Prefix, strings.ToLower(env.ModulePlacement(t, e).Step.Name))
		assert.Equal(t, prefix, outputs.String(t, "agc__agc_name"))
		assert.True(t, strings.HasSuffix(outputs.String(t, "agc__agc_id"), "/trafficControllers/"+prefix), "agc_id must be the AGC %s", prefix)
		assert.Equal(t, "private", outputs.String(t, "agc__frontend_name"))
		assert.True(t, strings.HasSuffix(outputs.String(t, "agc__frontend_fqdn"), ".privatelink.alb.azure.com"), "frontend_fqdn must be a private frontend")
		assert.NotEmpty(t, outputs.String(t, "agc__private_endpoint_ip"), "private_endpoint_ip was not returned")
		assert.Equal(t, netOutputs.StringList(t, "vpc__private_agc_subnet_cidrs")[0], outputs.String(t, "agc__subnet_cidr"), "subnet_cidr must be the vpc private-agc subnet")
	})
}
