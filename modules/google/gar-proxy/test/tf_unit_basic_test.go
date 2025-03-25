package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)

func TestTerraformGarProxy(t *testing.T) {
	t.Run("Biz", testTerraformGarProxyBiz)
	t.Run("Pri", testTerraformGarProxyPri)
}

func testTerraformGarProxyBiz(t *testing.T) {
	t.Parallel()
	outputs := google.GetTFOutputs(t, "biz")

	registry := fmt.Sprintf("biz-%s-gar-proxy", strings.ToLower(os.Getenv("STEP_NAME")))
	if len(registry) > 50 {
		registry = registry[:50]
	}

	hub_registry := tf.GetStringValue(t, outputs, "gar-proxy__hub_registry")
	ghcr_registry := tf.GetStringValue(t, outputs, "gar-proxy__ghcr_registry")
	gcr_registry := tf.GetStringValue(t, outputs, "gar-proxy__gcr_registry")
	ecr_registry := tf.GetStringValue(t, outputs, "gar-proxy__ecr_registry")
	quay_registry := tf.GetStringValue(t, outputs, "gar-proxy__quay_registry")
	k8s_registry := tf.GetStringValue(t, outputs, "gar-proxy__k8s_registry")

	assert.Equal(t, hub_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-hub", registry), "Wrong value for hub_registry")
	assert.Equal(t, ghcr_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-ghcr", registry), "Wrong value for ghcr_registry")
	assert.Equal(t, gcr_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-gcr", registry), "Wrong value for gcr_registry")
	assert.Equal(t, ecr_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-ecr", registry), "Wrong value for ecr_registry")
	assert.Equal(t, quay_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-quay", registry), "Wrong value for quay_registry")
	assert.Equal(t, k8s_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-k8s", registry), "Wrong value for k8s_registry")
}

func testTerraformGarProxyPri(t *testing.T) {
	t.Parallel()
	outputs := google.GetTFOutputs(t, "pri")

	registry := fmt.Sprintf("pri-%s-gar-proxy", strings.ToLower(os.Getenv("STEP_NAME")))
	if len(registry) > 50 {
		registry = registry[:50]
	}

	hub_registry := tf.GetStringValue(t, outputs, "gar-proxy__hub_registry")
	ghcr_registry := tf.GetStringValue(t, outputs, "gar-proxy__ghcr_registry")
	gcr_registry := tf.GetStringValue(t, outputs, "gar-proxy__gcr_registry")
	ecr_registry := tf.GetStringValue(t, outputs, "gar-proxy__ecr_registry")
	quay_registry := tf.GetStringValue(t, outputs, "gar-proxy__quay_registry")
	k8s_registry := tf.GetStringValue(t, outputs, "gar-proxy__k8s_registry")

	assert.Equal(t, hub_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-hub", registry), "Wrong value for hub_registry")
	assert.Equal(t, ghcr_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-ghcr", registry), "Wrong value for ghcr_registry")
	assert.Equal(t, gcr_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-gcr", registry), "Wrong value for gcr_registry")
	assert.Equal(t, ecr_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-ecr", registry), "Wrong value for ecr_registry")
	assert.Equal(t, quay_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-quay", registry), "Wrong value for quay_registry")
	assert.Equal(t, k8s_registry, fmt.Sprintf("europe-north1-docker.pkg.dev/entigo-infralib2/%s-k8s", registry), "Wrong value for k8s_registry")
}
