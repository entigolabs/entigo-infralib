package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformAgc(t *testing.T) {
	t.Run("Biz", testTerraformAgcBiz)
	t.Run("Pri", testTerraformAgcPri)
}

func checkAgc(t *testing.T, env string) {
	outputs := azure.GetTFOutputs(t, env)
	netOutputs := azure.GetTFOutputsStep(t, env, "net")

	prefix := fmt.Sprintf("%s-%s-agc", env, strings.ToLower(os.Getenv("STEP_NAME")))
	assert.Equal(t, prefix, tf.GetStringValue(t, outputs, "agc__agc_name"))
	assert.True(t, strings.HasSuffix(tf.GetStringValue(t, outputs, "agc__agc_id"), "/trafficControllers/"+prefix), "agc_id must be the AGC %s", prefix)
	assert.Equal(t, "private", tf.GetStringValue(t, outputs, "agc__frontend_name"))
	assert.True(t, strings.HasSuffix(tf.GetStringValue(t, outputs, "agc__frontend_fqdn"), ".privatelink.alb.azure.com"), "frontend_fqdn must be a private frontend")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "agc__private_endpoint_ip"), "private_endpoint_ip was not returned")
	assert.Equal(t, tf.GetStringListValue(t, netOutputs, "vpc__private_agc_subnet_cidrs")[0], tf.GetStringValue(t, outputs, "agc__subnet_cidr"), "subnet_cidr must be the vpc private-agc subnet")
}

func testTerraformAgcBiz(t *testing.T) {
	t.Parallel()
	checkAgc(t, "biz")
}

func testTerraformAgcPri(t *testing.T) {
	t.Parallel()
	checkAgc(t, "pri")
}
