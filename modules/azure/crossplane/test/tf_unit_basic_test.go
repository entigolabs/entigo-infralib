package test

import (
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

const guid = `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`

func TestTerraformCrossplane(t *testing.T) {
	t.Run("Biz", testTerraformCrossplaneBiz)
	t.Run("Pri", testTerraformCrossplanePri)
}

func checkCrossplane(t *testing.T, env string) {
	outputs := azure.GetTFOutputs(t, env)

	for _, key := range []string{"client_id", "principal_id", "core_client_id", "core_principal_id", "tenant_id"} {
		assert.Regexp(t, guid, tf.GetStringValue(t, outputs, "crossplane__"+key))
	}
	if tenant := strings.TrimSpace(os.Getenv("AZURE_TENANT_ID")); tenant != "" {
		assert.Equal(t, tenant, tf.GetStringValue(t, outputs, "crossplane__tenant_id"))
	}
	assert.Equal(t, azure.SubscriptionID(), tf.GetStringValue(t, outputs, "crossplane__subscription_id"))
	assert.Equal(t, azure.ResourceGroup(env), tf.GetStringValue(t, outputs, "crossplane__resource_group_name"))
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "crossplane__kubernetes_namespace"))
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "crossplane__kubernetes_service_account"))
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "crossplane__kubernetes_core_service_account"))
}

func testTerraformCrossplaneBiz(t *testing.T) {
	t.Parallel()
	checkCrossplane(t, "biz")
}

func testTerraformCrossplanePri(t *testing.T) {
	t.Parallel()
	checkCrossplane(t, "pri")
}
