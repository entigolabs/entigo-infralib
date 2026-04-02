package test

import (
	"fmt"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/require"
)

// applyFile applies a kubectl manifest file.
// The manifest itself specifies namespaces, so opts namespace is not required to match.
func applyFile(t *testing.T, opts *terrak8s.KubectlOptions, file string) {
	t.Helper()
	_, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "apply", "-f", file)
	require.NoError(t, err, "kubectl apply -f %s", file)
}

// syncWithRetry forces an ArgoCD sync and retries up to syncMaxAttempts times on errors.
//
// Why not check sync.status == "Synced":
// An app can be OutOfSync (resources drifted externally) but still healthy and working.
// Instead we check operationState.phase == "Succeeded" — did the last sync operation succeed.
func syncWithRetry(t *testing.T, opts *terrak8s.KubectlOptions, appName string) {
	t.Helper()
	for attempt := 1; attempt <= syncMaxAttempts; attempt++ {
		err := trySyncAndWait(t, opts, appName)
		if err == nil {
			return
		}
		if attempt == syncMaxAttempts {
			t.Fatalf("Application %q failed to sync after %d attempts: %v", appName, syncMaxAttempts, err)
		}
		t.Logf("[sync] attempt %d/%d for %q failed: %v — retrying in %s", attempt, syncMaxAttempts, appName, err, syncRetryDelay)
		time.Sleep(syncRetryDelay)
	}
}

// trySyncAndWait triggers one ArgoCD sync operation and waits for it to reach a terminal phase.
// Returns nil on Succeeded, non-nil on Failed/Error/timeout.
func trySyncAndWait(t *testing.T, opts *terrak8s.KubectlOptions, appName string) error {
	t.Helper()
	_, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "patch", "application", appName,
		"--type", "merge", "-p", `{"operation":{"initiatedBy":{"username":"test"},"sync":{}}}`)
	if err != nil {
		return fmt.Errorf("trigger sync: %w", err)
	}

	phase, err := retry.DoWithRetryE(t, fmt.Sprintf("sync operation %q", appName), syncPollRetries, syncPollInterval,
		func() (string, error) {
			p, err := getAppField(t, opts, appName, ".status.operationState.phase")
			if err != nil {
				return "", err
			}
			// Succeeded/Failed/Error are terminal; anything else (Running, "") keep polling
			if p == "Succeeded" || p == "Failed" || p == "Error" {
				return p, nil
			}
			return "", fmt.Errorf("phase=%q", p)
		})
	if err != nil {
		return fmt.Errorf("waiting for sync operation: %w", err)
	}
	if phase != "Succeeded" {
		msg, _ := getAppField(t, opts, appName, ".status.operationState.message")
		return fmt.Errorf("sync %s: %s", phase, msg)
	}
	return nil
}

// waitApplicationHealthy waits until ArgoCD reports the Application as Healthy.
// Healthy means all managed resources are in a working state, regardless of sync status.
func waitApplicationHealthy(t *testing.T, opts *terrak8s.KubectlOptions, appName string) {
	t.Helper()
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("Application %q Healthy", appName), healthRetries, healthInterval,
		func() (string, error) {
			health, err := getAppField(t, opts, appName, ".status.health.status")
			if err != nil {
				return "", err
			}
			if health != "Healthy" {
				return "", fmt.Errorf("health=%q", health)
			}
			return health, nil
		})
	require.NoError(t, err, "Application %q never became Healthy", appName)
}

func getAppField(t *testing.T, opts *terrak8s.KubectlOptions, appName, jsonPath string) (string, error) {
	t.Helper()
	return terrak8s.RunKubectlAndGetOutputE(t, opts, "get", "application", appName,
		"-o", fmt.Sprintf("jsonpath={%s}", jsonPath))
}
