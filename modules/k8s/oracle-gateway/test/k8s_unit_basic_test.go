package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestOracleGateway(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		// Only the internal gateway is deployed, see test/oracle_biz.yaml.
		_, err := k8s.WaitUntilK8SGatewayAvailable(t, c, "internal", 50, 6*time.Second)
		require.NoError(t, err, "oracle-gateway not available error")
	})
}
