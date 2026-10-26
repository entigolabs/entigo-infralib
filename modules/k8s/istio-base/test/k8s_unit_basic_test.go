package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/k8s"
)

func TestIstioBase(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		c := k8s.Connect(t, e)

		require.NoError(t, k8s.WaitUntilResourcesAvailable(t, c, "networking.istio.io/v1beta1", []string{"virtualservices"}, 60, 1*time.Second), "Istio Base no VirtualService CRD")
		require.NoError(t, k8s.WaitUntilResourcesAvailable(t, c, "networking.istio.io/v1beta1", []string{"gateways"}, 60, 1*time.Second), "Istio Base no Gateway CRD")
	})
}
