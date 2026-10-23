package test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/azure"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestAksNodePool(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		infraOutputs := tf.GetStep(t, e, "infra")

		prefix := fmt.Sprintf("%s-%s-aks-node-pool", e.Prefix, strings.ToLower(env.ModulePlacement(t, e).Step.Name))
		poolName := regexp.MustCompile("[^a-z0-9]").ReplaceAllString(prefix, "")[:12]
		assert.Equal(t, poolName, outputs.String(t, "aks-node-pool__node_pool_name"))
		assert.NotEmpty(t, outputs.String(t, "aks-node-pool__node_pool_id"), "node_pool_id was not returned")
		assert.Equal(t, infraOutputs.String(t, "aks__cluster_id"), outputs.String(t, "aks-node-pool__cluster_id"), "cluster_id must be the infra aks cluster")
	})
}
