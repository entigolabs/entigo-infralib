package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

// ── Wait helpers ─────────────────────────────────────────────────────────────

// waitSyncedAndReady polls until a Crossplane resource has Synced=True and Ready=True.
func waitSyncedAndReady(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string, retries int, interval time.Duration) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("%s/%s Synced+Ready", kind, name), retries, interval,
		func() (string, error) { return checkConditions(t, opts, kind, name, "Synced", "Ready") })
	require.NoError(t, err)
}

// waitSyncedAndReadyByLabel waits for a Crossplane resource identified by crossplane.io/composite label.
// Returns the resource name once ready.
func waitSyncedAndReadyByLabel(t *testing.T, opts *terrak8s.KubectlOptions, kind, composite string, retries int, interval time.Duration) string {
	t.Helper()
	var name string
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("%s composite=%s Synced+Ready", kind, composite), retries, interval,
		func() (string, error) {
			n, err := getFirstByLabel(t, opts, kind, composite)
			if err != nil || n == "" {
				return "", fmt.Errorf("no %s with composite=%s: %v", kind, composite, err)
			}
			result, err := checkConditions(t, opts, kind, n, "Synced", "Ready")
			if err == nil {
				name = n
			}
			return result, err
		})
	require.NoError(t, err)
	return name
}

// waitCrossplanePackageReady waits for a Crossplane Function or Configuration to be Healthy+Installed.
func waitCrossplanePackageReady(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("%s/%s Healthy+Installed", kind, name), 40, 6*time.Second,
		func() (string, error) { return checkConditions(t, opts, kind, name, "Healthy", "Installed") })
	require.NoError(t, err)
}

// waitResourceExists polls until a named resource exists.
func waitResourceExists(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string, retries int, interval time.Duration) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("%s/%s exists", kind, name), retries, interval,
		func() (string, error) {
			n, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "-o", "jsonpath={.metadata.name}")
			if err != nil || n == "" {
				return "", fmt.Errorf("%s/%s not found", kind, name)
			}
			return n, nil
		})
	require.NoError(t, err)
}

// waitFieldEquals polls until a jsonpath field on a resource equals the expected value.
func waitFieldEquals(t *testing.T, opts *terrak8s.KubectlOptions, kind, name, fieldPath, expected string, retries int, interval time.Duration) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("%s/%s %s==%s", kind, name, fieldPath, expected), retries, interval,
		func() (string, error) {
			val, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "-o", fmt.Sprintf("jsonpath={%s}", fieldPath))
			if err != nil {
				return "", err
			}
			if val != expected {
				return "", fmt.Errorf("got %q, want %q", val, expected)
			}
			return val, nil
		})
	require.NoError(t, err, "%s/%s: field %s never reached %q", kind, name, fieldPath, expected)
}

// ── Patch helpers ─────────────────────────────────────────────────────────────

// patchResource applies a JSON merge patch to a resource.
func patchResource(t *testing.T, opts *terrak8s.KubectlOptions, kind, name, patch string) {
	t.Helper()
	_, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "patch", kind, name, "--type", "merge", "-p", patch)
	require.NoError(t, err, "patch %s/%s", kind, name)
}

// patchAndWaitField patches a resource and waits for a field on the SAME resource to reach a value.
// For cross-resource propagation (e.g. patch CR, wait on provider-managed resource), use patchResource + waitFieldEquals.
func patchAndWaitField(t *testing.T, opts *terrak8s.KubectlOptions, kind, name, patch, fieldPath, expected string, retries int, interval time.Duration) {
	t.Helper()
	patchResource(t, opts, kind, name, patch)
	waitFieldEquals(t, opts, kind, name, fieldPath, expected, retries, interval)
}

// ── Read helpers ──────────────────────────────────────────────────────────────

// getField reads a single jsonpath field from a resource. Fails the test if the resource is not found.
func getField(t *testing.T, opts *terrak8s.KubectlOptions, kind, name, fieldPath string) string {
	t.Helper()
	val, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "-o", fmt.Sprintf("jsonpath={%s}", fieldPath))
	require.NoError(t, err, "get %s from %s/%s", fieldPath, kind, name)
	return val
}

// getFirstByLabel returns the name of the first resource matching crossplane.io/composite=<composite>.
func getFirstByLabel(t *testing.T, opts *terrak8s.KubectlOptions, kind, composite string) (string, error) {
	t.Helper()
	out, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind,
		"-l", fmt.Sprintf("crossplane.io/composite=%s", composite),
		"-o", "jsonpath={.items[0].metadata.name}")
	return strings.TrimSpace(out), err
}

// ── Delete protection helpers ─────────────────────────────────────────────────

// testDeletionRejected verifies that a delete is rejected at the API level (e.g. by a validating webhook).
func testDeletionRejected(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string) {
	t.Helper()
	out, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "delete", kind, name, "--wait=false")
	combined := out
	if err != nil {
		combined += err.Error()
	}
	require.Error(t, err, "deletion of %s/%s should be rejected by webhook", kind, name)
	require.Contains(t, combined, "protected", "rejection message should mention protection, got: %s", combined)
}

// testUsageBlocksDeletion verifies that a Crossplane Usage resource protects kind/name from deletion.
// It attempts deletion and then confirms the resource still exists after a brief wait.
func testUsageBlocksDeletion(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string) {
	t.Helper()
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, opts, "delete", kind, name, "--wait=false")
	time.Sleep(10 * time.Second)
	existing, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
	require.NoError(t, err)
	require.Equal(t, name, existing, "%s/%s was deleted despite Usage protection", kind, name)
}

// ── Usage verification ────────────────────────────────────────────────────────

// testUsage waits for a Usage resource to exist and verifies its of/by fields and replayDeletion flag.
func testUsage(t *testing.T, opts *terrak8s.KubectlOptions, usageName, ofKind, ofName, byKind, byName string) {
	t.Helper()
	waitResourceExists(t, opts, UsageKind, usageName, 30, 10*time.Second)
	require.Equal(t, ofKind, getField(t, opts, UsageKind, usageName, ".spec.of.kind"))
	require.Equal(t, ofName, getField(t, opts, UsageKind, usageName, ".spec.of.resourceRef.name"))
	require.Equal(t, byKind, getField(t, opts, UsageKind, usageName, ".spec.by.kind"))
	require.Equal(t, byName, getField(t, opts, UsageKind, usageName, ".spec.by.resourceRef.name"))
	require.Equal(t, "true", getField(t, opts, UsageKind, usageName, ".spec.replayDeletion"))
}

// ── Cleanup helpers ───────────────────────────────────────────────────────────

// cleanupDeleteParallel deletes multiple resources of the same kind concurrently and waits for all to disappear.
func cleanupDeleteParallel(t *testing.T, opts *terrak8s.KubectlOptions, kind string, names ...string) {
	if len(names) == 0 {
		return
	}
	for _, name := range names {
		_, _ = terrak8s.RunKubectlAndGetOutputE(t, opts, "delete", kind, name,
			"--cascade=foreground", "--wait=false", "--ignore-not-found")
	}
	var wg sync.WaitGroup
	for _, name := range names {
		name := name
		wg.Add(1)
		go func() {
			defer wg.Done()
			cleanupWaitGone(t, opts, kind, name, 30)
		}()
	}
	wg.Wait()
}

// cleanupDeleteAndWait deletes a single resource and waits for it to disappear.
func cleanupDeleteAndWait(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string, maxRetries int) {
	_, _ = terrak8s.RunKubectlAndGetOutputE(t, opts, "delete", kind, name,
		"--cascade=foreground", "--wait=false", "--ignore-not-found")
	cleanupWaitGone(t, opts, kind, name, maxRetries)
}

func cleanupWaitGone(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string, maxRetries int) {
	_, _ = retry.DoWithRetryE(t, fmt.Sprintf("waiting for %s/%s deletion", kind, name), maxRetries, 10*time.Second,
		func() (string, error) {
			out, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "--ignore-not-found", "-o", "jsonpath={.metadata.name}")
			if err != nil {
				return "", err
			}
			if out != "" {
				return "", fmt.Errorf("%s/%s still exists", kind, name)
			}
			return "deleted", nil
		})
}

// ── Internal ──────────────────────────────────────────────────────────────────

func checkConditions(t *testing.T, opts *terrak8s.KubectlOptions, kind, name string, conditions ...string) (string, error) {
	t.Helper()
	for _, cond := range conditions {
		status, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "get", kind, name, "-o",
			fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].status}`, cond))
		if err != nil {
			return "", err
		}
		if status != "True" {
			return "", fmt.Errorf("%s/%s: %s=%s", kind, name, cond, status)
		}
	}
	return strings.Join(conditions, "+"), nil
}
