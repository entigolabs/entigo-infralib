package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformKms(t *testing.T) {
	t.Run("Biz", testTerraformServicesBiz)
}

func testTerraformServicesBiz(t *testing.T) {
	t.Parallel()
	outputs := google.GetTFOutputs(t, "biz")
	value := tf.GetValue(t, outputs, "services__service_agent_emails")
	services, ok := value.(map[string]interface{})
	require.True(t, ok, "services__service_agent_emails is not a map")

	expected := map[string]string{
		"artifactregistry.googleapis.com":   "service-394873127837@gcp-sa-artifactregistry.iam.gserviceaccount.com",
		"certificatemanager.googleapis.com": "service-394873127837@gcp-sa-certificatemanager.iam.gserviceaccount.com",
		"clouddeploy.googleapis.com":        "service-394873127837@gcp-sa-clouddeploy.iam.gserviceaccount.com",
		"cloudkms.googleapis.com":           "service-394873127837@gcp-sa-ekms.iam.gserviceaccount.com",
		"cloudscheduler.googleapis.com":     "service-394873127837@gcp-sa-cloudscheduler.iam.gserviceaccount.com",
		"compute.googleapis.com":            "service-394873127837@compute-system.iam.gserviceaccount.com",
		"container.googleapis.com":          "service-394873127837@container-engine-robot.iam.gserviceaccount.com",
		"dns.googleapis.com":                "service-394873127837@gcp-sa-dns.iam.gserviceaccount.com",
		"file.googleapis.com":               "service-394873127837@cloud-filer.iam.gserviceaccount.com",
		"memorystore.googleapis.com":        "service-394873127837@gcp-sa-memorystore.iam.gserviceaccount.com",
		"pubsub.googleapis.com":             "service-394873127837@gcp-sa-pubsub.iam.gserviceaccount.com",
		"redis.googleapis.com":              "service-394873127837@cloud-redis.iam.gserviceaccount.com",
		"run.googleapis.com":                "service-394873127837@serverless-robot-prod.iam.gserviceaccount.com",
		"secretmanager.googleapis.com":      "service-394873127837@gcp-sa-secretmanager.iam.gserviceaccount.com",
		"servicenetworking.googleapis.com":  "service-394873127837@service-networking.iam.gserviceaccount.com",
		"sqladmin.googleapis.com":           "service-394873127837@gcp-sa-cloud-sql.iam.gserviceaccount.com",
		"storage.googleapis.com":            "service-394873127837@gs-project-accounts.iam.gserviceaccount.com",
	}

	assert.Len(t, services, len(expected))
	for key, expectedEmail := range expected {
		actualEmail, exists := services[key]
		assert.True(t, exists, "missing service: %s", key)
		assert.Equal(t, expectedEmail, actualEmail, "mismatch for service: %s", key)
	}
}
