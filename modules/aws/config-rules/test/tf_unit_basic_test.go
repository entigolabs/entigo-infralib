package test

import (
	"os"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"

	terraaws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformConfigRules(t *testing.T) {
	t.Run("Biz", testTerraformConfigRulesBiz)
}

func testTerraformConfigRulesBiz(t *testing.T) {
	t.Parallel()

	awsRegion := terraaws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	outputs := aws.GetTFOutputs(t, "biz")

	configRulesLogsBucketName := tf.GetStringValue(t, outputs, "config-rules__config_rules_logs_bucket_name")
	assert.NotEmpty(t, configRulesLogsBucketName, "config_rules_logs_bucket_name must not be empty")

	err := aws.WaitUntilAWSBucketExists(t, awsRegion, configRulesLogsBucketName, 30, 4*time.Second)
	require.NoError(t, err, "S3 bucket creation error")

	err = aws.WaitUntilBucketFileAvailable(t, configRulesLogsBucketName, "config-logs/AWSLogs/877483565445/Config/ConfigWritabilityCheckFile", 20, 6*time.Second)
	if err != nil {
		t.Fatal("File not found in AWS bucket:", err)
	}
}
