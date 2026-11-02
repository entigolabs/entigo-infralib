package test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

const guid = `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`

func TestCrossplane(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		for _, key := range []string{"client_id", "principal_id", "core_client_id", "core_principal_id", "tenant_id"} {
			assert.Regexp(t, guid, outputs.String(t, "crossplane__"+key))
		}
		// The framework has no tenant accessor; the credentials of the executor carry it.
		if tenant := strings.TrimSpace(os.Getenv("AZURE_TENANT_ID")); tenant != "" {
			assert.Equal(t, tenant, outputs.String(t, "crossplane__tenant_id"))
		}
		assert.Equal(t, e.MustSubscriptionID(t), outputs.String(t, "crossplane__subscription_id"))
		assert.Equal(t, azure.ResourceGroup(t, e), outputs.String(t, "crossplane__resource_group_name"))
		assert.NotEmpty(t, outputs.String(t, "crossplane__kubernetes_namespace"))
		assert.NotEmpty(t, outputs.String(t, "crossplane__kubernetes_service_account"))
		assert.NotEmpty(t, outputs.String(t, "crossplane__kubernetes_core_service_account"))
	})
}
