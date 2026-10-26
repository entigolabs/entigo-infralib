package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestEcrProxy(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		registryName := fmt.Sprintf("%s-%s-ecr-proxy", e.Prefix, strings.ToLower(env.ModulePlacement(t, e).Step.Name))
		if len(registryName) > 24 {
			registryName = registryName[:24]
		}
		registryHost := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", aws.AccountID(t, e), e.MustRegion(t))

		ecrRegistry := outputs.String(t, "ecr-proxy__ecr_registry")
		assert.Equal(t, ecrRegistry, fmt.Sprintf("%s/%s-ecr", registryHost, registryName), "No correct value for ecr_registry")

		ghcrRegistry := outputs.String(t, "ecr-proxy__ghcr_registry")
		assert.Equal(t, ghcrRegistry, fmt.Sprintf("%s/%s-ghcr", registryHost, registryName), "No correct value for ghcr_registry")

		hubRegistry := outputs.String(t, "ecr-proxy__hub_registry")
		assert.Equal(t, hubRegistry, fmt.Sprintf("%s/%s-hub", registryHost, registryName), "No correct value for hub_registry")

		k8sRegistry := outputs.String(t, "ecr-proxy__k8s_registry")
		assert.Equal(t, k8sRegistry, fmt.Sprintf("%s/%s-k8s", registryHost, registryName), "No correct value for k8s_registry")

		quayRegistry := outputs.String(t, "ecr-proxy__quay_registry")
		assert.Equal(t, quayRegistry, fmt.Sprintf("%s/%s-quay", registryHost, registryName), "No correct value for quay_registry")
	})
}
