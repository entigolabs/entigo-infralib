package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/env"
	_ "github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestGarProxy(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		// Repository ids are the module prefix cut to 50 characters, plus the upstream's suffix.
		registry := fmt.Sprintf("%s-%s-gar-proxy", e.Prefix, env.ModulePlacement(t, e).Step.Name)
		if len(registry) > 50 {
			registry = registry[:50]
		}
		host := fmt.Sprintf("%s-docker.pkg.dev/%s", e.MustRegion(t), e.MustProject(t))

		for _, upstream := range []string{"hub", "ghcr", "gcr", "ecr", "quay", "k8s"} {
			assert.Equal(t, fmt.Sprintf("%s/%s-%s", host, registry, upstream), outputs.String(t, "gar-proxy__"+upstream+"_registry"), "Wrong value for %s_registry", upstream)
		}
	})
}
