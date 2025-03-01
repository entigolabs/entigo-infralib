package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"context"

	"cloud.google.com/go/storage"
	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
)

func GetTFOutputs(t testing.TestingT, prefix string) map[string]interface{} {
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	return GetTFOutputsStep(t, prefix, stepName)
}

func GetTFOutputsStep(t testing.TestingT, prefix string, stepName string) map[string]interface{} {
	Region := gcp.GetRandomRegion(t, os.Getenv("GOOGLE_PROJECT"), []string{os.Getenv("GOOGLE_REGION")}, nil)
	bucket := fmt.Sprintf("%s-%s-%s", prefix, os.Getenv("GOOGLE_PROJECT"), Region)
	file := fmt.Sprintf("%s-%s/terraform-output.json", prefix, stepName)

	reader, err := gcp.ReadBucketObjectE(t, bucket, file)
	require.NoError(t, err, "Failed to get module outputs region %s bucket %s file %s Error: %s", Region, bucket, file, err)

	outputs, err := io.ReadAll(reader)
	require.NoError(t, err, "Failed to read object contents: %v", err)

	// Close the reader
	if closer, ok := reader.(io.Closer); ok {
		defer closer.Close()
	}

	//fmt.Printf("OUTPUT %s %s", file, string(outputs))
	var result map[string]interface{}
	err = json.Unmarshal(outputs, &result)
	require.NoError(t, err, "Error parsing JSON: %s Error: %s", string(outputs), err)
	return result
}

func WaitUntilGCPBucketExists(t testing.TestingT, name string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for bucket %s to be created", name)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		err := gcp.AssertStorageBucketExistsE(t, name)
		if err != nil {
			return "", err
		}
		return "Bucket is now available", nil
	},
	)
	if err != nil {
		logger.Log(t, "Timed out waiting for bucket to be created: %s", err)
		return err
	}
	logger.Log(t, message)
	return nil
}

func WaitUntilGCPBucketDeleted(t testing.TestingT, name string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for bucket %s to be deleted", name)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		err := gcp.AssertStorageBucketExistsE(t, name)
		if err != nil {
			if errors.Is(err, storage.ErrBucketNotExist) {
				return "Bucket is now deleted", nil
			}
			return "", err
		}
		return "", fmt.Errorf("bucket still exists")
	},
	)
	if err != nil {
		logger.Log(t, "Timed out waiting for bucket to be deleted: %s", err)
		return err
	}
	logger.Log(t, message)
	return nil
}

func WaitUntilBucketFileAvailable(t testing.TestingT, bucket string, file string, retries int, sleepBetweenRetries time.Duration) error {
	statusMsg := fmt.Sprintf("Wait for bucket %s file %s", bucket, file)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		_, err := gcp.ReadBucketObjectE(t, bucket, file)
		if err != nil {
			return "", err
		}
		return "Bucket is now available", nil
	},
	)
	if err != nil {
		logger.Log(t, "Timed out waiting for bucket to be created: %s", err)
		return err
	}
	logger.Log(t, message)
	return nil
}


func WaitUntilGoogleCloudDnsRecordExists(t testing.TestingT, zoneName, recordName, recordType string, retries int, sleepBetweenRetries time.Duration) error {
	projectID := os.Getenv("GOOGLE_PROJECT")
	message := fmt.Sprintf("Checking if DNS record %s (%s) exists in zone %s", recordName, recordType, zoneName)

	_, err := retry.DoWithRetryE(t, message, retries, sleepBetweenRetries, func() (string, error) {
		exists, err := CheckDNSRecordExists(projectID, zoneName, recordName, recordType)
		if err != nil {
			return "", err
		}
		if !exists {
			return "", fmt.Errorf("DNS record %s (%s) not found yet", recordName, recordType)
		}
		return "DNS record found", nil
	})

	if err != nil {
		t.Fatalf("Failed to find DNS record %s (%s) after %d retries: %v", recordName, recordType, retries, err)
		return err
	}
	return nil
}

// CheckDNSRecordExists checks if a DNS record exists in a specific zone.
func CheckDNSRecordExists(projectID, zoneName, recordName, recordType string) (bool, error) {
	ctx := context.Background()
	dnsService, err := dns.NewService(ctx, option.WithScopes(dns.NdevClouddnsReadwriteScope))
	if err != nil {
		return false, fmt.Errorf("failed to create DNS service: %w", err)
	}

	resp, err := dnsService.ResourceRecordSets.List(projectID, zoneName).Name(fmt.Sprintf("%s.",recordName)).Type(recordType).Do()
	if err != nil {
		return false, fmt.Errorf("failed to list DNS records: %w", err)
	}

	return len(resp.Rrsets) > 0, nil
}
