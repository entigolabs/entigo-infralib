package test

import (
	"strings"
	"testing"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

const (
	// zoneAContributorGroup is mapped only to the contributor role of zone a (test zone fixture).
	zoneAContributorGroup = "zone-a-contributor"
	// zoneMaintainerGroup is mapped only to the maintainer role (environment roleMapping).
	zoneMaintainerGroup = "maintainer"
)

type rbacCheck struct {
	verb, resource string
	want           bool
}

// testZoneRBAC asks the API server (kubectl auth can-i with impersonation) what maintainers and
// contributors may do in a zone namespace. Requests for the Namespace object are sent with the
// namespace set, as the API server authorizes them that way (kubernetes/kubernetes#142024).
func testZoneRBAC(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Run("zone-rbac", func(t *testing.T) {
		ensureKyvernoTestNamespace(t, cluster)
		waitNamespaceRoleBinding(t, cluster, KyvernoTestNSName, "contributor")
		waitNamespaceRoleBinding(t, cluster, KyvernoTestNSName, "maintainer")
		if t.Failed() {
			return
		}
		ns := "namespaces/" + KyvernoTestNSName

		shared := []rbacCheck{
			{"create", "deployments.apps", true},
			{"delete", "pods", true},
			{"create", "pods/exec", true},
			{"get", "pods/log", true},
			{"create", "secrets", true},
			{"update", "configmaps", true},
			{"create", "cronjobs.batch", true},
			{"create", "ingresses.networking.k8s.io", true},
			{"create", "virtualservices.networking.istio.io", true},
			{"create", "serviceentries.networking.istio.io", true},
			{"create", "externalsecrets.external-secrets.io", true},
			{"create", "postgresqlinstances.database.entigo.com", true},
			{"create", "webapps.workload.entigo.com", true},
			{"create", "rolepolicyattachments.iam.aws.m.upbound.io", false},
			{"create", "roles.iam.aws.m.upbound.io", false},
			{"create", "buckets.s3.aws.m.upbound.io", false},
			{"create", "rolebindings.rbac.authorization.k8s.io", false},
			{"create", "roles.rbac.authorization.k8s.io", false},
			{"delete", "networkpolicies.networking.k8s.io", false},
			{"update", "sidecars.networking.istio.io", false},
			{"create", "authorizationpolicies.security.istio.io", false},
			{"create", "policyexceptions.policies.kyverno.io", false},
			{"update", "resourcequotas", false},
		}

		t.Run("contributor", func(t *testing.T) {
			t.Parallel()
			checks := append([]rbacCheck{
				{"update", ns, false},
				{"patch", ns, false},
				{"delete", ns, false},
			}, shared...)
			assertRBAC(t, cluster, zoneAContributorGroup, checks)
		})
		t.Run("maintainer", func(t *testing.T) {
			t.Parallel()
			checks := append([]rbacCheck{
				{"update", ns, true},
				{"patch", ns, true},
				{"delete", ns, true},
			}, shared...)
			assertRBAC(t, cluster, zoneMaintainerGroup, checks)
		})
	})
}

func assertRBAC(t *testing.T, cluster *terrak8s.KubectlOptions, group string, checks []rbacCheck) {
	t.Helper()
	for _, c := range checks {
		got := canI(t, cluster, group, c.verb, c.resource)
		if got != c.want {
			t.Errorf("group %s: can-i %s %s in %s = %v, want %v", group, c.verb, c.resource, KyvernoTestNSName, got, c.want)
		}
	}
}

// canI returns the API server's answer for the group; kubectl exits non-zero on "no".
func canI(t *testing.T, cluster *terrak8s.KubectlOptions, group, verb, resource string) bool {
	t.Helper()
	nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KyvernoTestNSName)
	out, _ := terrak8s.RunKubectlAndGetOutputE(t, nsOpts, "auth", "can-i", verb, resource,
		"--as", "zone-rbac-test", "--as-group", group)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	answer := strings.TrimSpace(lines[len(lines)-1])
	switch {
	case answer == "yes":
		return true
	case strings.HasPrefix(answer, "no"):
		return false
	default:
		t.Fatalf("unexpected can-i output for %s %s: %q", verb, resource, out)
		return false
	}
}
