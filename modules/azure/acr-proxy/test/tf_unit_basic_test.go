package test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestAcrProxy(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"azure_biz": testAcrProxyBiz,
		"azure_pri": testAcrProxyPri,
	})
}

func acrName(t *testing.T, e *env.Environment) string {
	stepName := strings.TrimSpace(strings.ToLower(env.ModulePlacement(t, e).Step.Name))
	name := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(fmt.Sprintf("%s-%s-acr-proxy", e.Prefix, stepName), "")
	if len(name) > 42 {
		name = name[:42]
	}
	return name + azure.UniqueSuffix(e.Prefix, e.MustSubscriptionID(t), e.MustRegion(t))
}

func checkRegistries(t *testing.T, e *env.Environment, outputs tf.Outputs) {
	loginServer := outputs.String(t, "acr-proxy__acr_login_server")
	assert.Equal(t, fmt.Sprintf("%s.azurecr.io", acrName(t, e)), loginServer)
	assert.NotEmpty(t, outputs.String(t, "acr-proxy__acr_id"), "acr_id was not returned")
	for registry, upstream := range map[string]string{
		"hub":  "docker.io",
		"ghcr": "ghcr.io",
		"gcr":  "gcr.io",
		"ecr":  "public.ecr.aws",
		"quay": "quay.io",
		"mcr":  "mcr.microsoft.com",
		"k8s":  "registry.k8s.io",
	} {
		assert.Equal(t, fmt.Sprintf("%s/%s", loginServer, upstream), outputs.String(t, fmt.Sprintf("acr-proxy__%s_registry", registry)))
	}
}

// biz runs the Premium SKU with a private endpoint and the AKS bootstrap cache rule.
func testAcrProxyBiz(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	checkRegistries(t, e, outputs)
	assert.NotEmpty(t, outputs.String(t, "acr-proxy__private_endpoint_id"), "private_endpoint_id was not returned")
	assert.NotEmpty(t, outputs.String(t, "acr-proxy__aks_bootstrap_cache_rule_id"), "aks_bootstrap_cache_rule_id was not returned")
}

// pri runs the Standard SKU, which has neither.
func testAcrProxyPri(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	checkRegistries(t, e, outputs)
	assert.False(t, outputs.HasKeyWithPrefix("acr-proxy__private_endpoint_id"), "private_endpoint_id must be null on Standard")
	assert.False(t, outputs.HasKeyWithPrefix("acr-proxy__aks_bootstrap_cache_rule_id"), "aks_bootstrap_cache_rule_id must be null")
}
