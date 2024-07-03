package google

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gruntwork-io/terratest/modules/gcp"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

func SetupBucket(t testing.TestingT, bucketName string) string {
	Region := gcp.GetRandomRegion(t, os.Getenv("GOOGLE_PROJECT"), []string{os.Getenv("GOOGLE_REGION")}, nil)
	err := gcp.CreateStorageBucketE(t, os.Getenv("GOOGLE_PROJECT"), bucketName, nil)
	if err != nil {
		if strings.Contains(err.Error(), "Your previous request to create the named bucket succeeded and you already own it.") {
			logger.Log(t, "Bucket already owned by you. Skipping bucket creation.")
		} else {
			t.Fatal(err)
		}
	}
	err = WaitUntilGCPBucketExists(t, bucketName, 30, 2*time.Second)
	require.NoError(t, err, "Bucket creation error")
	return Region
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

func GetProjectID(t testing.TestingT) string {
	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		logger.Logf(t, "Failed to find default credentials: %v", err)
	}
	crmService, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(credentials))
	if err != nil {
		logger.Logf(t, "Failed to create cloudresourcemanager service: %v", err)
	}
	projectID := credentials.ProjectID
	if projectID == "" {
		// If ProjectID is empty, fetch the list of projects
		projectListCall := crmService.Projects.List()
		projectList, err := projectListCall.Do()
		if err != nil {
			logger.Logf(t, "Failed to list projects: %v", err)
		}
		if len(projectList.Projects) > 0 {
			projectID = projectList.Projects[0].ProjectId
		}
	}
	return projectID
}

func GetSecret(t testing.TestingT, secretName string) string {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		logger.Logf(t, "failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	request := &secretmanagerpb.AccessSecretVersionRequest{Name: secretName}

	result, err := client.AccessSecretVersion(ctx, request)
	if err != nil {
		logger.Logf(t, "failed to access secret %v", err)
	}

	fmt.Printf("retrieved payload for: %s %s\n", result.Name, result.Payload.Data)

	secret := strings.Trim(strings.Split(fmt.Sprintf("%s", result.Payload.Data), ",")[0], `"`)

	return secret
}
