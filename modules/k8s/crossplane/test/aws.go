package test

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"time"
)

func WaitUntilBucketAvailable(t testing.TestingT, region string, name string, retries int, sleepBetweenRetries time.Duration) error {
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

func WaitUntilBucketDeleted(t testing.TestingT, region string, name string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for bucket %s in %s region to be deleted", name, region)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		err := aws.AssertS3BucketExistsE(t, region, name)
		if err != nil {
			var awsErr awserr.Error
			if errors.As(err, &awsErr) {
				if awsErr.Code() == s3.ErrCodeNoSuchBucket {
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
