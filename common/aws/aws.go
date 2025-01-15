package aws

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"encoding/json"
	
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
)

func GetTFOutputs (t testing.TestingT, prefix string, step string) map[string]interface{} {
        awsRegion := aws.GetRandomRegion(t, []string{os.Getenv("AWS_REGION")}, nil)
	bucket := fmt.Sprintf("%s-877483565445-eu-north-1", prefix)
	file := fmt.Sprintf("%s-%s/terraform-output.json", prefix, step)
        outputs, err := aws.GetS3ObjectContentsE(t, awsRegion, bucket, file)

	require.NoError(t, err, "Failed to get module outputs region %s bucket %s prefix %s Error: %s", awsRegion, bucket, file, err)
	fmt.Printf("%s %s", prefix, outputs)
	var result map[string]interface{}
	err = json.Unmarshal([]byte(outputs), &result)
	require.NoError(t, err, "Error parsing JSON: %s Error: %s", outputs, err)
	return result
}

func GetValue(t testing.TestingT, outputs map[string]interface{}, key string) (interface{}) {
    output, ok := outputs[key].(map[string]interface{})
    require.True(t, ok, "Error finding key %s in JSON %s", key, outputs)
    value, exists := output["value"]
    require.True(t, exists, "Error finding value %s from JSON %s Error", key, outputs)
    // Return the value as is - it could be a string, list, or any other type
    return value
}

func GetStringValue(t testing.TestingT, outputs map[string]interface{}, key string) (string) {
    value := GetValue(t, outputs, key)
    strValue, ok := value.(string)
    require.True(t, ok, "Fond value %s for %s is not a string", value, key)  
    return strValue
}

func GetStringListValue(t testing.TestingT, outputs map[string]interface{}, key string) ([]string) {
    value := GetValue(t, outputs, key)

    if listValue, ok := value.([]interface{}); ok {
        result := make([]string, len(listValue))
        for i, v := range listValue {
            if strValue, ok := v.(string); ok {
                result[i] = strValue
            } else {
	        logger.Logf(t, "value at index %d for key %s is not a string", i, key)
                return nil
            }
        }
        return result
    }
    logger.Logf(t, "value for key %s is not a list", key)
    return nil
}

func HasKeyWithPrefix(t testing.TestingT, outputs map[string]interface{}, prefix string) (bool) {

    for key := range outputs {
        if strings.HasPrefix(key, prefix) {
            return true
        }
    }

    return false
}


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
