package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/env"
	_ "github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestServices(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		services := outputs.Map(t, "services__service_agent_emails")

		// Service agents are named after the project number, not the project id.
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
	})
}
