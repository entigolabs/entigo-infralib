package test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

// getFirstByLabel returns the name of the first resource matching crossplane.io/composite label,
// or empty string if none exist. Uses items[*] to avoid a kubectl error on empty lists.
func getFirstByLabel(t *testing.T, opts *terrak8s.KubectlOptions, kind, composite string) (string, error) {
	t.Helper()
	out, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, "-l",
		fmt.Sprintf("crossplane.io/composite=%s", composite), "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return "", err
	}
	items := strings.Fields(out)
	if len(items) == 0 {
		return "", nil
	}
	return items[0], nil
}

func waitSyncedAndReady(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string, retries int, interval time.Duration) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s/%s", kind, name), retries, interval, func() (string, error) {
		for _, condType := range []string{"Synced", "Ready"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("%s/%s: %s=%s", kind, name, condType, status)
			}
		}
		return "Synced+Ready", nil
	})
	require.NoError(t, err)
}

func waitSyncedAndReadyByLabel(t *testing.T, opts *terrak8s.KubectlOptions, kind, composite string, retries int, interval time.Duration) string {
	t.Helper()
	var name string
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s (composite=%s)", kind, composite), retries, interval, func() (string, error) {
		out, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, "-l",
			fmt.Sprintf("crossplane.io/composite=%s", composite), "-o", "jsonpath={.items[*].metadata.name}")
		if err != nil {
			return "", err
		}
		items := strings.Fields(out)
		if len(items) == 0 {
			return "", fmt.Errorf("no %s found for composite=%s", kind, composite)
		}
		n := items[0]
		for _, condType := range []string{"Synced", "Ready"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, n, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("%s/%s: %s=%s", kind, n, condType, status)
			}
		}
		name = n
		return "Synced+Ready", nil
	})
	require.NoError(t, err)
	return name
}

func forceSyncArgoApp(t *testing.T, opts *terrak8s.KubectlOptions, appName string) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("force syncing ArgoCD app '%s'", appName), 30, 10*time.Second, func() (string, error) {
		_, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "patch", "application", appName, "--type", "merge", "-p",
			`{"operation":{"initiatedBy":{"username":"test"},"sync":{"revision":"HEAD"}}}`)
		return "", err
	})
	require.NoError(t, err, fmt.Sprintf("failed to force sync ArgoCD app '%s'", appName))
}

func waitArgoCDAppSyncedAndHealthy(t *testing.T, opts *terrak8s.KubectlOptions, appName string, retries int, interval time.Duration) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for ArgoCD app '%s'", appName), retries, interval, func() (string, error) {
		syncStatus, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "application", appName, "-o", "jsonpath={.status.sync.status}")
		if err != nil {
			return "", err
		}
		if syncStatus != "Synced" {
			return "", fmt.Errorf("app '%s' sync status: %s", appName, syncStatus)
		}
		healthStatus, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "application", appName, "-o", "jsonpath={.status.health.status}")
		if err != nil {
			return "", err
		}
		if healthStatus != "Healthy" {
			return "", fmt.Errorf("app '%s' health status: %s", appName, healthStatus)
		}
		return "Synced+Healthy", nil
	})
	require.NoError(t, err)
}

// setupRoleOptions creates kubectl options authenticated as an IAM user mapped to a Kubernetes
// role group. Credentials are read from the given environment variable names. If either env var
// is unset the test is skipped. The returned options share the same cluster as baseOptions.
func setupRoleOptions(t *testing.T, baseOptions *terrak8s.KubectlOptions, accessKeyIDEnv, secretAccessKeyEnv string) *terrak8s.KubectlOptions {
	t.Helper()
	accessKeyID := os.Getenv(accessKeyIDEnv)
	secretAccessKey := os.Getenv(secretAccessKeyEnv)
	if accessKeyID == "" || secretAccessKey == "" {
		t.Skipf("%s or %s not set", accessKeyIDEnv, secretAccessKeyEnv)
	}

	// EKS context ARN format: arn:aws:eks:REGION:ACCOUNT:cluster/CLUSTER-NAME
	parts := strings.Split(baseOptions.ContextName, ":")
	require.True(t, len(parts) >= 6, "unexpected EKS context ARN format: %s", baseOptions.ContextName)
	region := parts[3]
	clusterName := strings.TrimPrefix(parts[5], "cluster/")

	kubeconfigFile, err := os.CreateTemp("", "role-kubeconfig-*.yaml")
	require.NoError(t, err)
	kubeconfigPath := kubeconfigFile.Name()
	kubeconfigFile.Close()
	t.Cleanup(func() { os.Remove(kubeconfigPath) })

	cmd := exec.Command("aws", "eks", "update-kubeconfig",
		"--name", clusterName,
		"--region", region,
		"--kubeconfig", kubeconfigPath)
	cmd.Env = append(os.Environ(),
		"AWS_ACCESS_KEY_ID="+accessKeyID,
		"AWS_SECRET_ACCESS_KEY="+secretAccessKey,
	)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to create role kubeconfig: %s", output)

	out, err := exec.Command("kubectl", "config", "current-context", "--kubeconfig", kubeconfigPath).Output()
	require.NoError(t, err, "failed to get current context from role kubeconfig")

	return terrak8s.NewKubectlOptions(strings.TrimSpace(string(out)), kubeconfigPath, "")
}

func waitCrossplanePackageReady(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s/%s", kind, name), 40, 6*time.Second, func() (string, error) {
		for _, condType := range []string{"Healthy", "Installed"} {
			status, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "-o",
				fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, condType))
			if err != nil {
				return "", err
			}
			if status != "True" {
				return "", fmt.Errorf("%s/%s: %s=%s", kind, name, condType, status)
			}
		}
		return "Healthy+Installed", nil
	})
	require.NoError(t, err)
}
