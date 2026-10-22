package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/entigolabs/entigo-infralib-test/env"
	_ "github.com/entigolabs/entigo-infralib-test/google"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestCrossplane(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)

		serviceAccountEmail := outputs.String(t, "crossplane__service_account_email")
		projectId := outputs.String(t, "crossplane__project_id")

		assert.NotEmpty(t, projectId, "project_id was not returned")

		// The service account is named after the module prefix <prefix>-<step>-crossplane,
		// cut to Google's 28 character account id limit without a trailing hyphen.
		stepName := env.ModulePlacement(t, e).Step.Name
		googleServiceAccountId := strings.TrimRight(truncateString(fmt.Sprintf("%s-%s-crossplane", e.Prefix, stepName), 28), "-")
		googleServiceAccountEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", googleServiceAccountId, projectId)

		assert.Equal(t, googleServiceAccountEmail, serviceAccountEmail, "Wrong service_account_email returned")
	})
}

func truncateString(input string, maxLength int) string {
	if len(input) > maxLength {
		return input[:maxLength]
	}
	return input
}
