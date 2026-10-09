package azure

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/stretchr/testify/require"
)

func SubscriptionID() string {
	return strings.TrimSpace(os.Getenv("AZURE_SUBSCRIPTION_ID"))
}

func Location() string {
	if value := strings.TrimSpace(os.Getenv("AZURE_LOCATION")); value != "" {
		return value
	}
	return "westeurope"
}

func ResourceGroup(prefix string) string {
	return fmt.Sprintf("%s-infralib-%s", prefix, Location())
}

func UniqueSuffix(prefix string, subscriptionID string, location string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{prefix, subscriptionID, location}, "/")))
	x := binary.BigEndian.Uint64(sum[:8])
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	out := make([]byte, 8)
	for i := range out {
		out[i] = digits[x%36]
		x /= 36
	}
	return string(out)
}

func GetCredential(t testing.TestingT) azcore.TokenCredential {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	require.NoError(t, err, "Failed to create Azure credential: %s", err)
	return cred
}

func GetTFOutputs(t testing.TestingT, prefix string) map[string]interface{} {
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	return GetTFOutputsStep(t, prefix, stepName)
}

func GetTFOutputsStep(t testing.TestingT, prefix string, stepName string) map[string]interface{} {
	ctx := context.Background()
	cred := GetCredential(t)
	resourceGroup := ResourceGroup(prefix)
	file := fmt.Sprintf("%s-%s/terraform-output.json", prefix, stepName)
	logger.Logf(t, "File %s", file)

	accounts, err := armstorage.NewAccountsClient(SubscriptionID(), cred, nil)
	require.NoError(t, err, "Failed to create storage accounts client: %s", err)

	var errs []error
	pager := accounts.NewListByResourceGroupPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		require.NoError(t, err, "Failed to list storage accounts in %s: %s", resourceGroup, err)
		for _, account := range page.Value {
			outputs, err := ReadBlob(ctx, cred, *account.Name, "tfstate", file)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", *account.Name, err))
				continue
			}
			var result map[string]interface{}
			err = json.Unmarshal(outputs, &result)
			require.NoError(t, err, "Error parsing JSON: %s Error: %s", string(outputs), err)
			return result
		}
	}
	require.FailNow(t, fmt.Sprintf("Failed to get module outputs resource group %s file %s", resourceGroup, file), "%v", errors.Join(errs...))
	return nil
}

func ReadBlob(ctx context.Context, cred azcore.TokenCredential, accountName string, container string, blob string) ([]byte, error) {
	client, err := azblob.NewClient(fmt.Sprintf("https://%s.blob.core.windows.net/", accountName), cred, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.DownloadStream(ctx, container, blob, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func WaitUntilDnsRecordExists(t testing.TestingT, resourceGroup, zoneName, recordName, recordType string, retries int, sleepBetweenRetries time.Duration) error {
	client, err := armdns.NewRecordSetsClient(SubscriptionID(), GetCredential(t), nil)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("Checking if DNS record %s (%s) exists in zone %s", recordName, recordType, zoneName)
	_, err = retry.DoWithRetryE(t, message, retries, sleepBetweenRetries, func() (string, error) {
		_, err := client.Get(context.Background(), resourceGroup, zoneName, recordName, armdns.RecordType(recordType), nil)
		if err != nil {
			return "", err
		}
		return "DNS record found", nil
	})
	return err
}

func WaitUntilBlobAvailable(t testing.TestingT, account string, container string, blob string, retries int, sleepBetweenRetries time.Duration) error {
	cred := GetCredential(t)
	statusMsg := fmt.Sprintf("Wait for storage account %s container %s blob %s", account, container, blob)
	message, err := retry.DoWithRetryE(t, statusMsg, retries, sleepBetweenRetries, func() (string, error) {
		if _, err := ReadBlob(context.Background(), cred, account, container, blob); err != nil {
			return "", err
		}
		return "Blob is now available", nil
	})
	if err != nil {
		logger.Logf(t, "Timed out waiting for blob: %s", err)
		return err
	}
	logger.Log(t, message)
	return nil
}
