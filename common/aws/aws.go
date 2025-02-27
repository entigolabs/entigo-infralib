package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"os"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
)

func GetTFOutputs(t testing.TestingT, prefix string) map[string]interface{} {
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	bucket := fmt.Sprintf("%s-877483565445-%s", prefix, awsRegion)
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))

	file := fmt.Sprintf("%s-%s/terraform-output.json", prefix, stepName)
	logger.Logf(t, "File %s", file)
	outputs, err := aws.GetS3ObjectContentsE(t, awsRegion, bucket, file)

	require.NoError(t, err, "Failed to get module outputs region %s bucket %s file %s Error: %s", awsRegion, bucket, file, err)
	//fmt.Printf("OUTPUT %s %s", file, outputs)
	var result map[string]interface{}
	err = json.Unmarshal([]byte(outputs), &result)
	require.NoError(t, err, "Error parsing JSON: %s Error: %s", outputs, err)
	return result
}

func WaitUntilAWSBucketExists(t testing.TestingT, region string, name string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for bucket %s in %s region to be created", name, region)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		err := aws.AssertS3BucketExistsE(t, region, name)
		if err != nil {
			return "", err
		}
		return "Bucket is now available", nil
	},
	)
	if err != nil {
		logger.Logf(t, "Timed out waiting for bucket to be created: %s", err)
		return err
	}
	logger.Logf(t, message)
	return nil
}

func WaitUntilAWSBucketDeleted(t testing.TestingT, region string, name string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for bucket %s in %s region to be deleted", name, region)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		err := aws.AssertS3BucketExistsE(t, region, name)
		if err != nil {
			var awsErr awserr.Error
			if errors.As(err, &awsErr) {
				if awsErr.Code() == "NotFound" {
					return "Bucket is now deleted", nil
				}
			}
			return "", err
		}
		return "", fmt.Errorf("bucket still exists")
	},
	)
	if err != nil {
		logger.Logf(t, "Timed out waiting for bucket to be deleted: %s", err)
		return err
	}
	logger.Logf(t, message)
	return nil
}

func WaitUntilBucketFileAvailable(t testing.TestingT, bucket string, file string, retries int, sleepBetweenRetries time.Duration) error {
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	statusMsg := fmt.Sprintf("Wait for bucket %s file %s", bucket, file)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		_, err := aws.GetS3ObjectContentsE(t, awsRegion, bucket, file)
		if err != nil {
			return "", err
		}
		return "File is now available", nil
	},
	)
	if err != nil {
		logger.Logf(t, "Timed out waiting for bucket file to be created: %s", err)
		return err
	}
	logger.Logf(t, message)
	return nil
}

func WaitUntilAWSRoute53RecordExists(t testing.TestingT, hostedZoneID, recordName, recordType, awsRegion string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for Route53Record %s in %s region to be created", recordName, awsRegion)

	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		_, err := aws.GetRoute53RecordE(t, hostedZoneID, recordName, recordType, awsRegion)
		if err != nil {
			return "", err
		}
		return "Record is now available", nil
	})

	if err != nil {
		logger.Logf(t, "Timed out waiting for Route53Record to be created: %s", err)
		return err
	}

	logger.Logf(t, message)
	return nil
}
