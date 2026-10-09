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

func TestTerraformAks(t *testing.T) {
	t.Run("Biz", testTerraformAksBiz)
	t.Run("Pri", testTerraformAksPri)
}

func checkAks(t *testing.T, env string) map[string]interface{} {
	outputs := azure.GetTFOutputs(t, env)
	netOutputs := azure.GetTFOutputsStep(t, env, "net")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	prefix := fmt.Sprintf("%s-%s-aks", env, stepName)

	assert.Equal(t, prefix, tf.GetStringValue(t, outputs, "aks__cluster_name"))
	assert.Equal(t, azure.Location(), tf.GetStringValue(t, outputs, "aks__region"))
	assert.Equal(t, azure.ResourceGroup(env), tf.GetStringValue(t, outputs, "aks__resource_group_name"))
	assert.Equal(t, fmt.Sprintf("%s-nodes-%s", prefix, azure.Location()), tf.GetStringValue(t, outputs, "aks__node_resource_group"))
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "aks__node_resource_group_id"), "node_resource_group_id was not returned")
	assert.Equal(t, tf.GetStringListValue(t, netOutputs, "vpc__private_subnets")[0], tf.GetStringValue(t, outputs, "aks__vnet_subnet_id"), "vnet_subnet_id must be the first private subnet of net")
	assert.True(t, strings.HasPrefix(tf.GetStringValue(t, outputs, "aks__oidc_issuer_url"), "https://"), "oidc_issuer_url must be https")
	assert.True(t, strings.HasPrefix(tf.GetStringValue(t, outputs, "aks__cluster_endpoint"), "https://"), "cluster_endpoint must be https")
	assert.True(t, strings.HasPrefix(tf.GetStringValue(t, outputs, "aks__kubernetes_version"), "1.36"), "Wrong kubernetes_version")
	assert.Equal(t, "10.244.0.0/16", tf.GetStringValue(t, outputs, "aks__pod_cidr"))
	assert.Equal(t, "10.96.0.0/16", tf.GetStringValue(t, outputs, "aks__service_cidr"))
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "aks__kubelet_identity_client_id"), "kubelet_identity_client_id was not returned")
	assert.Regexp(t, `^[0-9a-f-]{36}$`, tf.GetStringValue(t, outputs, "aks__alb_controller_client_id"))
	return outputs
}

func testTerraformAksBiz(t *testing.T) {
	t.Parallel()
	outputs := checkAks(t, "biz")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "aks__disk_encryption_set_id"), "biz has kms, disk_encryption_set_id was not returned")
}

func testTerraformAksPri(t *testing.T) {
	t.Parallel()
	outputs := checkAks(t, "pri")
	assert.Equal(t, "", tf.GetStringValue(t, outputs, "aks__disk_encryption_set_id"), "pri has no kms")
}
