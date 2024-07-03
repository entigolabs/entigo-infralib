package aws

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
)

func SetupBucket(t testing.TestingT, bucketName string) string {
	awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	err := aws.CreateS3BucketE(t, awsRegion, bucketName)
	if err != nil {
		if strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
			logger.Log(t, "Bucket already owned by you. Skipping bucket creation.")
		} else {
			t.Fatal(err)
		}
	}
	err = WaitUntilAWSBucketExists(t, awsRegion, bucketName, 30, 2*time.Second)
	require.NoError(t, err, "Bucket creation error")
	return awsRegion
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
