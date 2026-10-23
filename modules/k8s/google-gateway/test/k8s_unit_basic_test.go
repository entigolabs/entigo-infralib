package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestGoogleGateway(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		_, err := k8s.WaitUntilK8SGatewayAvailable(t, c, fmt.Sprintf("%s-external", c.Namespace), 50, 6*time.Second)
		require.NoError(t, err, "google-gateway not available error")

		// Only biz creates the internal gateway as well.
		if e.Prefix == "biz" {
			_, err = k8s.WaitUntilK8SGatewayAvailable(t, c, fmt.Sprintf("%s-internal", c.Namespace), 50, 6*time.Second)
			require.NoError(t, err, "google-gateway not available error")
		}
	})
}
