package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformEcrProxy(t *testing.T) {
	t.Run("Biz", testTerraformEcrProxyBiz)
	t.Run("Pri", testTerraformEcrProxyPri)
}

func testTerraformEcrProxyBiz(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz")
	registry_name := fmt.Sprintf("biz-%s-ecr-proxy", strings.ToLower(os.Getenv("STEP_NAME")))
	if len(registry_name) > 24 {
		registry_name = registry_name[:24]
	}

	ecr_registry := tf.GetStringValue(t, outputs, "ecr-proxy__ecr_registry")
	assert.Equal(t, ecr_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-ecr", registry_name), "No correct value for ecr_registry")

	ghcr_registry := tf.GetStringValue(t, outputs, "ecr-proxy__ghcr_registry")
	assert.Equal(t, ghcr_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-ghcr", registry_name), "No correct value for ghcr_registry")

	hub_registry := tf.GetStringValue(t, outputs, "ecr-proxy__hub_registry")
	assert.Equal(t, hub_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-hub", registry_name), "No correct value for hub_registry")

	k8s_registry := tf.GetStringValue(t, outputs, "ecr-proxy__k8s_registry")
	assert.Equal(t, k8s_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-k8s", registry_name), "No correct value for k8s_registry")

	quay_registry := tf.GetStringValue(t, outputs, "ecr-proxy__quay_registry")
	assert.Equal(t, quay_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-quay", registry_name), "No correct value for quay_registry")
}

func testTerraformEcrProxyPri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri")
	registry_name := fmt.Sprintf("pri-%s-ecr-proxy", strings.ToLower(os.Getenv("STEP_NAME")))
	if len(registry_name) > 24 {
		registry_name = registry_name[:24]
	}

	ecr_registry := tf.GetStringValue(t, outputs, "ecr-proxy__ecr_registry")
	assert.Equal(t, ecr_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-ecr", registry_name), "No correct value for ecr_registry")

	ghcr_registry := tf.GetStringValue(t, outputs, "ecr-proxy__ghcr_registry")
	assert.Equal(t, ghcr_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-ghcr", registry_name), "No correct value for ghcr_registry")

	hub_registry := tf.GetStringValue(t, outputs, "ecr-proxy__hub_registry")
	assert.Equal(t, hub_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-hub", registry_name), "No correct value for hub_registry")

	k8s_registry := tf.GetStringValue(t, outputs, "ecr-proxy__k8s_registry")
	assert.Equal(t, k8s_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-k8s", registry_name), "No correct value for k8s_registry")

	quay_registry := tf.GetStringValue(t, outputs, "ecr-proxy__quay_registry")
	assert.Equal(t, quay_registry, fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-quay", registry_name), "No correct value for quay_registry")
}
