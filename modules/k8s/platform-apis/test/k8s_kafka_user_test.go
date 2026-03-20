package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

const (
	AwsSecretKind    = "secret.secretsmanager.aws.m.upbound.io"
	SecretVersionKind = "secretversion.secretsmanager.aws.m.upbound.io"
	SecretPolicyKind  = "secretpolicy.secretsmanager.aws.m.upbound.io"
	ScramAssocKind    = "singlescramsecretassociation.kafka.aws.m.upbound.io"
	AclKind           = "accesscontrollist.acl.kafka.m.crossplane.io"
)

func runKafkaUserTests(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	// KafkaUser has sequenced phases, needs longer timeout
	waitSyncedAndReady(t, namespaceOptions, KafkaUserKind, KafkaUserName, 120, 10*time.Second)

	t.Run("sub-resources", func(t *testing.T) {
		t.Run("k8s-secret", func(t *testing.T) {
			t.Parallel()
			testKafkaUserK8sSecretExists(t, namespaceOptions)
		})
		t.Run("aws-secret", func(t *testing.T) {
			t.Parallel()
			testKafkaUserAwsSecretSyncedAndReady(t, namespaceOptions)
		})
		t.Run("secret-version", func(t *testing.T) {
			t.Parallel()
			testKafkaUserSecretVersionSyncedAndReady(t, namespaceOptions)
		})
		t.Run("secret-policy", func(t *testing.T) {
			t.Parallel()
			testKafkaUserSecretPolicySyncedAndReady(t, namespaceOptions)
		})
		t.Run("scram-association", func(t *testing.T) {
			t.Parallel()
			testKafkaUserScramAssociationSyncedAndReady(t, namespaceOptions)
		})
		t.Run("acls", func(t *testing.T) {
			t.Parallel()
			testKafkaUserAclsSyncedAndReady(t, namespaceOptions)
		})
	})

	testKafkaUserK8sSecretFields(t, namespaceOptions)
}

func testKafkaUserK8sSecretExists(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	expectedName := fmt.Sprintf("%s-%s-k8s", KafkaClusterName, KafkaUserName)
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for k8s Secret '%s'", expectedName), 60, 10*time.Second, func() (string, error) {
		name, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", "secret", expectedName, "-o", "jsonpath={.metadata.name}")
		if err != nil {
			return "", err
		}
		if name == "" {
			return "", fmt.Errorf("k8s Secret '%s' not found", expectedName)
		}
		return name, nil
	})
	require.NoError(t, err, "k8s Secret '%s' not found", expectedName)
}

func testKafkaUserAwsSecretSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	expectedName := fmt.Sprintf("%s-%s", KafkaClusterName, KafkaUserName)
	waitSyncedAndReadyByLabel(t, namespaceOptions, AwsSecretKind, KafkaUserName, 60, 10*time.Second)

	// Verify the AWS secret name field
	awsSecretName, err := getFirstByLabel(t, namespaceOptions, AwsSecretKind, KafkaUserName)
	require.NoError(t, err, "failed to find AWS Secret")
	require.NotEmpty(t, awsSecretName, "no AWS Secret found for composite '%s'", KafkaUserName)

	nameField, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", AwsSecretKind, awsSecretName, "-o", "jsonpath={.spec.forProvider.name}")
	require.NoError(t, err, "failed to get AWS Secret name field")
	expectedAWSName := fmt.Sprintf("AmazonMSK_%s-%s", KafkaClusterName, KafkaUserName)
	require.Equal(t, expectedAWSName, nameField, "AWS Secret name mismatch, expected '%s', resource '%s'", expectedAWSName, expectedName)
}

func testKafkaUserSecretVersionSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	waitSyncedAndReadyByLabel(t, namespaceOptions, SecretVersionKind, KafkaUserName, 60, 10*time.Second)
}

func testKafkaUserSecretPolicySyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	waitSyncedAndReadyByLabel(t, namespaceOptions, SecretPolicyKind, KafkaUserName, 60, 10*time.Second)
}

func testKafkaUserScramAssociationSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	waitSyncedAndReadyByLabel(t, namespaceOptions, ScramAssocKind, KafkaUserName, 60, 10*time.Second)
}

func testKafkaUserAclsSyncedAndReady(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for ACLs (composite=%s)", KafkaUserName), 60, 10*time.Second, func() (string, error) {
		aclNames, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", AclKind,
			"-l", fmt.Sprintf("crossplane.io/composite=%s", KafkaUserName), "-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			return "", err
		}
		names := strings.Fields(aclNames)
		// Expect: 1 consumer group ACL (test-group) + 2 topic ACLs (Read + Write) = 3
		if len(names) < 3 {
			return "", fmt.Errorf("expected at least 3 ACLs for composite=%s, found %d", KafkaUserName, len(names))
		}
		for _, name := range names {
			for _, condType := range []string{"Synced", "Ready"} {
				status, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", AclKind, name, "-o",
					fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
				if err != nil {
					return "", err
				}
				if status != "True" {
					return "", fmt.Errorf("ACL '%s': %s=%s", name, condType, status)
				}
			}
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err, "ACLs for KafkaUser '%s' failed to become Synced and Ready", KafkaUserName)
}

func testKafkaUserK8sSecretFields(t *testing.T, namespaceOptions *terrak8s.KubectlOptions) {
	t.Helper()
	secretName := fmt.Sprintf("%s-%s-k8s", KafkaClusterName, KafkaUserName)

	// Verify secretString data key exists
	data, err := terrak8s.RunKubectlAndGetOutputE(t, namespaceOptions, "get", "secret", secretName, "-o", "jsonpath={.data.secretString}")
	require.NoError(t, err, "failed to get secretString from k8s Secret '%s'", secretName)
	require.NotEmpty(t, data, "k8s Secret '%s' secretString is empty", secretName)
}
