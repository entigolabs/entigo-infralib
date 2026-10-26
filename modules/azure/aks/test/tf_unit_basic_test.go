package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestAks(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"azure_biz": testAksBiz,
		"azure_pri": testAksPri,
	})
}

func checkAks(t *testing.T, e *env.Environment) tf.Outputs {
	outputs := tf.Get(t, e)
	netOutputs := tf.GetStep(t, e, "net")
	stepName := strings.TrimSpace(strings.ToLower(env.ModulePlacement(t, e).Step.Name))
	prefix := fmt.Sprintf("%s-%s-aks", e.Prefix, stepName)
	location := e.MustRegion(t)

	assert.Equal(t, prefix, outputs.String(t, "aks__cluster_name"))
	assert.Equal(t, location, outputs.String(t, "aks__region"))
	assert.Equal(t, azure.ResourceGroup(t, e), outputs.String(t, "aks__resource_group_name"))
	assert.Equal(t, fmt.Sprintf("%s-nodes-%s", prefix, location), outputs.String(t, "aks__node_resource_group"))
	assert.NotEmpty(t, outputs.String(t, "aks__node_resource_group_id"), "node_resource_group_id was not returned")
	assert.Equal(t, netOutputs.StringList(t, "vpc__private_subnets")[0], outputs.String(t, "aks__vnet_subnet_id"), "vnet_subnet_id must be the first private subnet of net")
	assert.True(t, strings.HasPrefix(outputs.String(t, "aks__cluster_oidc_issuer_url"), "https://"), "cluster_oidc_issuer_url must be https")
	assert.True(t, strings.HasPrefix(outputs.String(t, "aks__cluster_endpoint"), "https://"), "cluster_endpoint must be https")
	assert.True(t, strings.HasPrefix(outputs.String(t, "aks__kubernetes_version"), "1.36"), "Wrong kubernetes_version")
	assert.Equal(t, "10.244.0.0/16", outputs.String(t, "aks__pod_cidr"))
	assert.Equal(t, "10.96.0.0/16", outputs.String(t, "aks__service_cidr"))
	assert.NotEmpty(t, outputs.String(t, "aks__kubelet_identity_client_id"), "kubelet_identity_client_id was not returned")
	assert.Regexp(t, `^[0-9a-f-]{36}$`, outputs.String(t, "aks__alb_controller_client_id"))
	return outputs
}

// biz has the kms module, so the cluster disks use its disk encryption set.
func testAksBiz(t *testing.T, e *env.Environment) {
	outputs := checkAks(t, e)
	assert.NotEmpty(t, outputs.String(t, "aks__disk_encryption_set_id"), "biz has kms, disk_encryption_set_id was not returned")
}

func testAksPri(t *testing.T, e *env.Environment) {
	outputs := checkAks(t, e)
	assert.Equal(t, "", outputs.String(t, "aks__disk_encryption_set_id"), "pri has no kms")
}
