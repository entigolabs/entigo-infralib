package test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/azure"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformAcrProxy(t *testing.T) {
	t.Run("Biz", testTerraformAcrProxyBiz)
	t.Run("Pri", testTerraformAcrProxyPri)
}

func acrName(env string) string {
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	name := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(fmt.Sprintf("%s-%s-acr-proxy", env, stepName), "")
	if len(name) > 42 {
		name = name[:42]
	}
	return name + azure.UniqueSuffix(env, azure.SubscriptionID(), azure.Location())
}

func checkRegistries(t *testing.T, outputs map[string]interface{}, env string) {
	loginServer := tf.GetStringValue(t, outputs, "acr-proxy__acr_login_server")
	assert.Equal(t, fmt.Sprintf("%s.azurecr.io", acrName(env)), loginServer)
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "acr-proxy__acr_id"), "acr_id was not returned")
	for registry, upstream := range map[string]string{
		"hub":  "docker.io",
		"ghcr": "ghcr.io",
		"gcr":  "gcr.io",
		"ecr":  "public.ecr.aws",
		"quay": "quay.io",
		"mcr":  "mcr.microsoft.com",
		"k8s":  "registry.k8s.io",
	} {
		assert.Equal(t, fmt.Sprintf("%s/%s", loginServer, upstream), tf.GetStringValue(t, outputs, fmt.Sprintf("acr-proxy__%s_registry", registry)))
	}
}

func testTerraformAcrProxyBiz(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "biz")
	checkRegistries(t, outputs, "biz")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "acr-proxy__private_endpoint_id"), "private_endpoint_id was not returned")
	assert.NotEmpty(t, tf.GetStringValue(t, outputs, "acr-proxy__aks_bootstrap_cache_rule_id"), "aks_bootstrap_cache_rule_id was not returned")
}

func testTerraformAcrProxyPri(t *testing.T) {
	t.Parallel()
	outputs := azure.GetTFOutputs(t, "pri")
	checkRegistries(t, outputs, "pri")
	assert.False(t, tf.HasKeyWithPrefix(t, outputs, "acr-proxy__private_endpoint_id"), "private_endpoint_id must be null on Standard")
	assert.False(t, tf.HasKeyWithPrefix(t, outputs, "acr-proxy__aks_bootstrap_cache_rule_id"), "aks_bootstrap_cache_rule_id must be null")
}
