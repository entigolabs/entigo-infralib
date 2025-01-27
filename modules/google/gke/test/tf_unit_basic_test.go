package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/stretchr/testify/assert"
)


func TestTerraformGke(t *testing.T) {
	t.Run("Biz", testTerraformGkeBiz)
	t.Run("Pri", testTerraformGkePri)
}

func testTerraformGkeBiz(t *testing.T) {
	t.Parallel()
	testTerraformGke(t, "biz")
}

func testTerraformGkePri(t *testing.T) {
	t.Parallel()
	testTerraformGke(t, "pri")
}

func testTerraformGke(t *testing.T, envName string) {
	outputs := google.GetTFOutputs(t, envName, "infra")
	
	clusterId := tf.GetStringValue(t, outputs, "gke__cluster_id")
	assert.NotEmpty(t, clusterId, "cluster_id was not returned")
      
	clusterEndpoint := tf.GetStringValue(t, outputs, "gke__cluster_endpoint")
	assert.NotEmpty(t, clusterEndpoint, "cluster_endpoint was not returned")
	
	clusterName := tf.GetStringValue(t, outputs, "gke__cluster_name")
	assert.NotEmpty(t, clusterName, "cluster_name was not returned")
	
	region := tf.GetStringValue(t, outputs, "gke__region")
	assert.NotEmpty(t, region, "region was not returned")
	
	serviceAccount := tf.GetStringValue(t, outputs, "gke__service_account")
	assert.NotEmpty(t, serviceAccount, "service_account was not returned")

}
