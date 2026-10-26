package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestConfigRules(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		region := e.MustRegion(t)
		outputs := tf.Get(t, e)

		configRulesLogsBucketName := outputs.String(t, "config-rules__config_rules_logs_bucket_name")
		assert.NotEmpty(t, configRulesLogsBucketName, "config_rules_logs_bucket_name must not be empty")

		err := aws.WaitUntilBucketExists(t, region, configRulesLogsBucketName, 30, 4*time.Second)
		require.NoError(t, err, "S3 bucket creation error")

		// AWS Config writes this file to prove it can deliver to the bucket.
		err = aws.WaitUntilBucketFileAvailable(t, region, configRulesLogsBucketName, "config-logs/AWSLogs/"+aws.AccountID(t, e)+"/Config/ConfigWritabilityCheckFile", 20, 6*time.Second)
		require.NoError(t, err, "File not found in AWS bucket")
	})
}
