package test

import (
        "fmt"
	"os"
	"testing"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/stretchr/testify/assert"
)


func TestTerraformCrossplane(t *testing.T) {
	t.Run("Biz", testTerraformCrossplaneBiz)
	t.Run("Pri", testTerraformCrossplanePri)
}

func testTerraformCrossplaneBiz(t *testing.T) {
	t.Parallel()
	testTerraformCrossplane(t, "biz")
}

func testTerraformCrossplanePri(t *testing.T) {
	t.Parallel()
	testTerraformCrossplane(t, "pri")
}

func testTerraformCrossplane(t *testing.T, envName string) {
	outputs := google.GetTFOutputs(t, envName, "infra")
	
	serviceAccountEmail := tf.GetStringValue(t, outputs, "crossplane__service_account_email")
	projectId := tf.GetStringValue(t, outputs, "crossplane__project_id")
	
	assert.NotEmpty(t, projectId, "project_id was not returned")

	googleServiceAccountId := truncateString(fmt.Sprintf("%s-crossplane", envName), 28)
	if os.Getenv("STEP_NAME") != "runner-main" {
		googleServiceAccountId = truncateString(fmt.Sprintf("%s-%s-crossplane", envName, os.Getenv("STEP_NAME")), 28)
	}
	googleServiceAccountEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", googleServiceAccountId, projectId)

	assert.Equal(t, googleServiceAccountEmail, serviceAccountEmail, "Wrong service_account_email returned")

}

func truncateString(input string, maxLength int) string {
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}
