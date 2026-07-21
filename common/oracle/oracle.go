package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	ocicommon "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
	"github.com/stretchr/testify/require"
)

// configProvider resolves OCI credentials the same way the agent and CLI do:
// OCI_CONFIG_FILE if set, otherwise the default ~/.oci/config DEFAULT profile.
func configProvider() ocicommon.ConfigurationProvider {
	configFile := os.Getenv("OCI_CONFIG_FILE")
	if configFile != "" {
		return ocicommon.CustomProfileConfigProvider(configFile, "DEFAULT")
	}
	return ocicommon.DefaultConfigProvider()
}

func newClient(region string) (objectstorage.ObjectStorageClient, string, error) {
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(configProvider())
	if err != nil {
		return objectstorage.ObjectStorageClient{}, "", err
	}
	if region != "" {
		client.SetRegion(region)
	}
	namespace, err := client.GetNamespace(context.Background(), objectstorage.GetNamespaceRequest{})
	if err != nil {
		return objectstorage.ObjectStorageClient{}, "", fmt.Errorf("failed to get object storage namespace: %w", err)
	}
	return client, *namespace.Value, nil
}

func GetTFOutputs(t testing.TestingT, prefix string) map[string]interface{} {
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	return GetTFOutputsStep(t, prefix, stepName)
}

// GetTFOutputsStep reads the terraform-output.json written by the agent for the given
// step. Bucket/object naming mirrors oracle/config.go's getBucketName ("{prefix}-{region}")
// and the object path convention shared across all clouds ("{prefix}-{stepName}/terraform-output.json").
func GetTFOutputsStep(t testing.TestingT, prefix string, stepName string) map[string]interface{} {
	region := os.Getenv("OCI_REGION")
	bucket := fmt.Sprintf("%s-%s", prefix, region)
	file := fmt.Sprintf("%s-%s/terraform-output.json", prefix, stepName)

	content, err := GetBucketObjectE(region, bucket, file)
	require.NoError(t, err, "Failed to get module outputs region %s bucket %s file %s Error: %s", region, bucket, file, err)

	var result map[string]interface{}
	err = json.Unmarshal(content, &result)
	require.NoError(t, err, "Error parsing JSON: %s Error: %s", string(content), err)
	return result
}

func GetBucketObjectE(region, bucket, file string) ([]byte, error) {
	client, namespace, err := newClient(region)
	if err != nil {
		return nil, err
	}
	response, err := client.GetObject(context.Background(), objectstorage.GetObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &bucket,
		ObjectName:    &file,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Content.Close() }()
	return io.ReadAll(response.Content)
}

func BucketExistsE(region, bucket string) (bool, error) {
	client, namespace, err := newClient(region)
	if err != nil {
		return false, err
	}
	_, err = client.GetBucket(context.Background(), objectstorage.GetBucketRequest{
		NamespaceName: &namespace,
		BucketName:    &bucket,
	})
	if err == nil {
		return true, nil
	}
	if serviceErr, ok := ocicommon.IsServiceError(err); ok && serviceErr.GetHTTPStatusCode() == 404 {
		return false, nil
	}
	return false, err
}

func WaitUntilOCIBucketExists(t testing.TestingT, region string, name string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for bucket %s in %s region to be created", name, region)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		exists, err := BucketExistsE(region, name)
		if err != nil {
			return "", err
		}
		if !exists {
			return "", fmt.Errorf("bucket %s does not exist yet", name)
		}
		return "Bucket is now available", nil
	})
	if err != nil {
		logger.Log(t, "Timed out waiting for bucket to be created: %s", err)
		return err
	}
	logger.Log(t, message)
	return nil
}

func WaitUntilBucketFileAvailable(t testing.TestingT, region, bucket, file string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for bucket %s file %s", bucket, file)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		_, err := GetBucketObjectE(region, bucket, file)
		if err != nil {
			return "", err
		}
		return "File is now available", nil
	})
	if err != nil {
		logger.Log(t, "Timed out waiting for file to be created: %s", err)
		return err
	}
	logger.Log(t, message)
	return nil
}
