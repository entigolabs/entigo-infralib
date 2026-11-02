package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/env"
	_ "github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestGke(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		clusterId := outputs.String(t, "gke__cluster_id")
		assert.NotEmpty(t, clusterId, "cluster_id was not returned")

		clusterEndpoint := outputs.String(t, "gke__cluster_endpoint")
		assert.NotEmpty(t, clusterEndpoint, "cluster_endpoint was not returned")

		clusterName := outputs.String(t, "gke__cluster_name")
		assert.NotEmpty(t, clusterName, "cluster_name was not returned")

		region := outputs.String(t, "gke__region")
		assert.NotEmpty(t, region, "region was not returned")

		serviceAccount := outputs.String(t, "gke__service_account")
		assert.NotEmpty(t, serviceAccount, "service_account was not returned")
	})
}
