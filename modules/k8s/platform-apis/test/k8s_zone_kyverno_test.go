package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

// testZoneKyverno runs live Kyverno policy e2e tests that mirror the static policy tests in
// compositions/zone/test/zone_kyverno_policies_test.go, but exercise the real admission webhooks
// running in the cluster.
//
// Role-based tests (ContributorDeny, MaintainerNamespaceDeny, MaintainerInfralibZoneDeny) are
// skipped automatically when the corresponding AWS credential env vars are not set.
func testZoneKyverno(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Run("zone-kyverno", func(t *testing.T) {
		ensureKyvernoTestNamespace(t, cluster)
		if t.Failed() {
			return
		}

		contributorKeyID := os.Getenv(ContributorKeyIDEnv)
		contributorSecret := os.Getenv(ContributorSecretEnv)
		maintainerKeyID := os.Getenv(MaintainerKeyIDEnv)
		maintainerSecret := os.Getenv(MaintainerSecretEnv)

		t.Run("NamespacePodSecurity", func(t *testing.T) {
			t.Parallel()
			testKyvernoNamespacePodSecurity(t, cluster)
		})
		t.Run("ZoneDeletionCheck", func(t *testing.T) {
			t.Parallel()
			testKyvernoZoneDeletionCheck(t, cluster)
		})
		t.Run("ZoneNamespaceOwnership", func(t *testing.T) {
			t.Parallel()
			testKyvernoZoneNamespaceOwnership(t, cluster)
		})
		t.Run("GenerateNamespaceFromArgoApp", func(t *testing.T) {
			t.Parallel()
			testKyvernoGenerateNamespaceFromArgoApp(t, cluster)
		})
		if contributorKeyID != "" && contributorSecret != "" {
			t.Run("ContributorDeny", func(t *testing.T) {
				t.Parallel()
				contributor := roleKubectlOptions(t, cluster, contributorKeyID, contributorSecret)
				testKyvernoContributorDeny(t, contributor)
			})
		} else {
			t.Logf("SKIPPING ContributorDeny: %s OR %s ENV VARS NOT SET", ContributorKeyIDEnv, ContributorSecretEnv)
		}
		if maintainerKeyID != "" && maintainerSecret != "" {
			maintainer := roleKubectlOptions(t, cluster, maintainerKeyID, maintainerSecret)
			t.Run("MaintainerNamespaceDeny", func(t *testing.T) {
				t.Parallel()
				testKyvernoMaintainerNamespaceDeny(t, cluster, maintainer)
			})
			t.Run("MaintainerInfralibZoneDeny", func(t *testing.T) {
				t.Parallel()
				testKyvernoMaintainerInfralibZoneDeny(t, cluster, maintainer)
			})
		} else {
			t.Logf("SKIPPING MaintainerNamespaceDeny AND MaintainerInfralibZoneDeny: %s OR %s ENV VARS NOT SET", MaintainerKeyIDEnv, MaintainerSecretEnv)
		}
	})
}

// testKyvernoNamespacePodSecurity covers:
//   - platform-apis-zone-namespace-pod-security (ValidatingPolicy)
//   - platform-apis-namespace-add-missing-zone-label (MutatingPolicy)
func testKyvernoNamespacePodSecurity(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Run("pass: restricted enforce+warn when setting is restricted", func(t *testing.T) {
		t.Parallel()
		const name = "kyverno-psa-restricted"
		t.Cleanup(func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", name, "--ignore-not-found", "--wait=false")
		})
		_, err := kyvernoApply(t, cluster, nsYAML(t, kyvernoNsData{Name: name, Zone: ZoneAName, Enforce: "restricted", Warn: "restricted"}))
		assertKyvernoAllowed(t, err)
	})
	t.Run("fail: privileged enforce is denied when setting is restricted", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, cluster, nsYAML(t, kyvernoNsData{Name: "kyverno-psa-priv-enf", Zone: ZoneAName, Enforce: "privileged", Warn: "restricted"}))
		assertKyvernoDenied(t, out, err)
	})
	t.Run("fail: privileged warn is denied when setting is restricted", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, cluster, nsYAML(t, kyvernoNsData{Name: "kyverno-psa-priv-warn", Zone: ZoneAName, Enforce: "restricted", Warn: "privileged"}))
		assertKyvernoDenied(t, out, err)
	})
	t.Run("pass: baseline enforce+warn are allowed when setting is baseline", func(t *testing.T) {
		t.Parallel()
		const name = "kyverno-psa-baseline"
		t.Cleanup(func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", name, "--ignore-not-found", "--wait=false")
		})
		_, err := kyvernoApply(t, cluster, nsYAML(t, kyvernoNsData{Name: name, Zone: ZoneAName, Enforce: "baseline", Warn: "baseline"}))
		assertKyvernoAllowed(t, err)
	})
	t.Run("pass: namespace without zone label gets auto-assigned", func(t *testing.T) {
		t.Parallel()
		const name = "kyverno-no-zone-test"
		t.Cleanup(func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", name, "--ignore-not-found", "--wait=false")
		})
		_, err := kyvernoApply(t, cluster, nsYAML(t, kyvernoNsData{Name: name, Enforce: "restricted", Warn: "restricted"}))
		assertKyvernoAllowed(t, err)
	})
	t.Run("fail: zone label referencing non-existent zone is denied", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, cluster, nsYAML(t, kyvernoNsData{Name: "kyverno-bad-zone", Zone: "non-existent-zone-xyz", Enforce: "restricted", Warn: "restricted"}))
		assertKyvernoDenied(t, out, err)
	})
	t.Run("pass: system namespace kube-system is excluded", func(t *testing.T) {
		t.Parallel()
		_, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "apply", "-f",
			writeTempYAML(t, nsYAML(t, kyvernoNsData{Name: "kube-system", Enforce: "privileged", Warn: "privileged"})),
			"--dry-run=server")
		assertKyvernoAllowed(t, err)
	})
}

// testKyvernoContributorDeny covers platform-apis-zone-namespace-contributor-deny (ValidatingPolicy).
// Contributors are denied all namespace operations regardless of operation type.
// Uses real AWS credentials to authenticate as a contributor IAM identity.
func testKyvernoContributorDeny(t *testing.T, contributor *terrak8s.KubectlOptions) {
	t.Run("fail: contributor cannot create a namespace", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, contributor, nsYAML(t, kyvernoNsData{Name: "kyverno-contrib-create", Zone: ZoneAName, Enforce: "restricted", Warn: "restricted"}))
		assertForbidden(t, out, err)
	})
	t.Run("fail: contributor cannot update a namespace", func(t *testing.T) {
		t.Parallel()
		out, err := terrak8s.RunKubectlAndGetOutputE(t, contributor, "patch", "namespace", KyvernoTestNSName,
			"--type", "merge", "-p", `{"metadata":{"annotations":{"kyverno.io/test":"true"}}}`)
		assertForbidden(t, out, err)
	})
	t.Run("fail: contributor cannot delete a namespace", func(t *testing.T) {
		t.Parallel()
		out, err := terrak8s.RunKubectlAndGetOutputE(t, contributor, "delete", "namespace", KyvernoTestNSName)
		assertForbidden(t, out, err)
	})
}

// testKyvernoMaintainerNamespaceDeny covers platform-apis-zone-namespace-maintainer-deny (ValidatingPolicy).
// Maintainers cannot create or update namespaces carrying the infralib zone label.
func testKyvernoMaintainerNamespaceDeny(t *testing.T, cluster, maintainer *terrak8s.KubectlOptions) {
	t.Run("fail: maintainer cannot create namespace with infralib zone", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, maintainer, nsYAML(t, kyvernoNsData{Name: "kyverno-maint-infralib", Zone: "infralib", Enforce: "restricted", Warn: "restricted"}))
		assertKyvernoDenied(t, out, err)
	})
	t.Run("fail: maintainer cannot update namespace with infralib zone", func(t *testing.T) {
		t.Parallel()
		out, err := terrak8s.RunKubectlAndGetOutputE(t, maintainer, "patch", "namespace", KyvernoTestNSName,
			"--type", "merge", "-p", `{"metadata":{"labels":{"tenancy.entigo.com/zone":"infralib"}}}`)
		assertKyvernoDenied(t, out, err)
	})
	t.Run("pass: maintainer can create namespace without infralib zone", func(t *testing.T) {
		t.Parallel()
		const name = "kyverno-maint-ok"
		t.Cleanup(func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", name, "--ignore-not-found", "--wait=false")
		})
		_, err := kyvernoApply(t, maintainer, nsYAML(t, kyvernoNsData{Name: name, Zone: ZoneAName, Enforce: "restricted", Warn: "restricted"}))
		assertKyvernoAllowed(t, err)
	})
}

// testKyvernoMaintainerInfralibZoneDeny covers platform-apis-zone-maintainer-infralib-zone-deny (ValidatingPolicy).
// Maintainers cannot create or update the Zone named "infralib".
func testKyvernoMaintainerInfralibZoneDeny(t *testing.T, cluster, maintainer *terrak8s.KubectlOptions) {
	t.Run("fail: maintainer cannot create the infralib zone", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, maintainer, zoneYAML(t, kyvernoZoneData{Name: "infralib"}))
		assertKyvernoDenied(t, out, err)
	})
	t.Run("fail: maintainer cannot update the infralib zone", func(t *testing.T) {
		t.Parallel()
		existing, _ := terrak8s.RunKubectlAndGetOutputE(t, cluster, "get", ZoneKind, "infralib",
			"--ignore-not-found", "-o", "jsonpath={.metadata.name}")
		if strings.TrimSpace(existing) == "" {
			t.Skip("infralib zone does not exist in cluster, skipping UPDATE test")
		}
		out, err := terrak8s.RunKubectlAndGetOutputE(t, maintainer, "patch", ZoneKind, "infralib",
			"--type", "merge", "-p", `{"metadata":{"annotations":{"kyverno.io/test":"true"}}}`)
		assertKyvernoDenied(t, out, err)
	})
	t.Run("pass: maintainer can create a non-infralib zone", func(t *testing.T) {
		t.Parallel()
		const name = "kyverno-maint-zone"
		t.Cleanup(func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", ZoneKind, name, "--ignore-not-found", "--wait=false")
		})
		_, err := kyvernoApply(t, maintainer, zoneYAML(t, kyvernoZoneData{Name: name}))
		assertKyvernoAllowed(t, err)
	})
}

// testKyvernoZoneDeletionCheck covers platform-apis-zone-deletion-check-namespaces (ValidatingPolicy).
func testKyvernoZoneDeletionCheck(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Run("fail: zone deletion blocked when namespaces still attached", func(t *testing.T) {
		t.Parallel()
		out, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", ZoneKind, ZoneAName)
		assertKyvernoDenied(t, out, err)
	})
	t.Run("pass: zone deletion allowed when no namespaces attached", func(t *testing.T) {
		t.Parallel()
		const name = "kyverno-del-test"
		applyFile(t, cluster, writeTempYAML(t, zoneYAML(t, kyvernoZoneData{Name: name})))
		t.Cleanup(func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", ZoneKind, name, "--ignore-not-found", "--wait=false")
		})
		_, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", ZoneKind, name)
		assertKyvernoAllowed(t, err)
	})
}

// testKyvernoZoneNamespaceOwnership covers platform-apis-zone-namespace-ownership (ValidatingPolicy).
func testKyvernoZoneNamespaceOwnership(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Run("fail: cannot create default zone", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, cluster, zoneYAML(t, kyvernoZoneData{Name: "default"}))
		assertKyvernoDenied(t, out, err)
	})
	t.Run("fail: cannot claim namespace owned by another zone", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, cluster, zoneYAML(t, kyvernoZoneData{Name: "kyverno-owner-test", Namespaces: []string{KyvernoTestNSName}}))
		assertKyvernoDenied(t, out, err)
	})
	t.Run("fail: cannot claim namespace without zone label", func(t *testing.T) {
		t.Parallel()
		out, err := kyvernoApply(t, cluster, zoneYAML(t, kyvernoZoneData{Name: "kyverno-steal-test", Namespaces: []string{"kube-system"}}))
		assertKyvernoDenied(t, out, err)
	})
	t.Run("pass: zone can manage its own namespace", func(t *testing.T) {
		t.Parallel()
		original, _ := terrak8s.RunKubectlAndGetOutputE(t, cluster, "get", ZoneKind, ZoneAName,
			"-o", "jsonpath={.spec.namespaces}")
		t.Cleanup(func() {
			restore := strings.TrimSpace(original)
			if restore == "" || restore == "null" {
				restore = "null"
			}
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "patch", ZoneKind, ZoneAName,
				"--type", "merge", "-p", fmt.Sprintf(`{"spec":{"namespaces":%s}}`, restore))
		})
		_, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "patch", ZoneKind, ZoneAName,
			"--type", "merge", "-p", fmt.Sprintf(`{"spec":{"namespaces":[{"name":%q}]}}`, AAppsNamespace))
		assertKyvernoAllowed(t, err)
	})
}

// testKyvernoGenerateNamespaceFromArgoApp covers generate-namespace-from-argocd-app (GeneratingPolicy).
// Creates real ArgoCD Application objects and verifies Kyverno generates (or does not generate)
// the destination namespace based on the AppProject.
func testKyvernoGenerateNamespaceFromArgoApp(t *testing.T, cluster *terrak8s.KubectlOptions) {
	kyvernoNSOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KyvernoTestNSName)

	t.Run("pass: ArgoApp generates namespace", func(t *testing.T) {
		const (
			appName     = "kyverno-generate-test"
			generatedNS = "kyverno-generated-ns"
		)
		t.Cleanup(func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, kyvernoNSOpts, "delete", "application", appName, "--ignore-not-found", "--wait=false")
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", generatedNS, "--ignore-not-found", "--wait=false")
		})
		applyFile(t, cluster, writeTempYAML(t, argoAppYAML(t, kyvernoArgoAppData{
			Name: appName, Namespace: KyvernoTestNSName, DestNamespace: generatedNS, Project: "a",
		})))
		waitResourceExists(t, cluster, "namespace", generatedNS, 12, 5*time.Second)
	})

	t.Run("pass: ArgoApp with infralib project does not generate namespace", func(t *testing.T) {
		const (
			appName    = "kyverno-infralib-test"
			infralibNS = "kyverno-infralib-gen"
		)
		t.Cleanup(func() {
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, kyvernoNSOpts, "delete", "application", appName, "--ignore-not-found", "--wait=false")
			_, _ = terrak8s.RunKubectlAndGetOutputE(t, cluster, "delete", "namespace", infralibNS, "--ignore-not-found", "--wait=false")
		})
		applyFile(t, cluster, writeTempYAML(t, argoAppYAML(t, kyvernoArgoAppData{
			Name: appName, Namespace: KyvernoTestNSName, DestNamespace: infralibNS, Project: "infralib",
		})))
		time.Sleep(20 * time.Second)
		existing, _ := terrak8s.RunKubectlAndGetOutputE(t, cluster, "get", "namespace", infralibNS,
			"--ignore-not-found", "-o", "jsonpath={.metadata.name}")
		require.Empty(t, strings.TrimSpace(existing),
			"namespace %q should not be generated for infralib project app", infralibNS)
	})
}
