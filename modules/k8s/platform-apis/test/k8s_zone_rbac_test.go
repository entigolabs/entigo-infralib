package test

import (
	"encoding/json"
	"strings"
	"testing"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

const (
	// zoneAContributorGroup is mapped only to the contributor role of zone a (test zone fixture).
	zoneAContributorGroup = "zone-a-contributor"
	// zoneMaintainerGroup is mapped only to the maintainer role (environment roleMapping).
	zoneMaintainerGroup = "maintainer"
	// zoneBNamespace is a spec namespace of test zone b.
	zoneBNamespace = "b1"
)

// rbacCheck uses kubectl-like notation: "resource[/subresource][.group]", and "namespaces/<name>" for
// the Namespace object.
type rbacCheck struct {
	verb, resource string
	want           bool
}

// testZoneRBAC asks the API server (SubjectAccessReview) what maintainers and contributors may do in
// zone namespaces. Requests for the Namespace object carry the namespace itself, as the API server
// authorizes them that way (kubernetes/kubernetes#142024).
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
			// in the namespaceRole.rules chart defaults
			{"create", "certificates.cert-manager.io", true},
			{"create", "servicemonitors.monitoring.coreos.com", true},
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

		// a zone grant must not reach namespaces outside the zone
		outsideZone := func(namespace string) []rbacCheck {
			return []rbacCheck{
				{"update", "namespaces/" + namespace, false},
				{"patch", "namespaces/" + namespace, false},
				{"delete", "namespaces/" + namespace, false},
			}
		}
		// zone a contributors must not reach zone b (b1 is a namespace of zone b)
		otherZone := append([]rbacCheck{
			{"create", "deployments.apps", false},
			{"delete", "pods", false},
			{"create", "pods/exec", false},
			{"get", "secrets", false},
			{"update", "configmaps", false},
			{"create", "webapps.workload.entigo.com", false},
		}, outsideZone(zoneBNamespace)...)

		t.Run("contributor", func(t *testing.T) {
			t.Parallel()
			checks := append([]rbacCheck{
				{"update", ns, false},
				{"patch", ns, false},
				{"delete", ns, false},
			}, shared...)
			assertRBAC(t, cluster, zoneAContributorGroup, KyvernoTestNSName, checks)
			assertRBAC(t, cluster, zoneAContributorGroup, zoneBNamespace, otherZone)
			assertRBAC(t, cluster, zoneAContributorGroup, "kube-system", outsideZone("kube-system"))
		})
		t.Run("maintainer", func(t *testing.T) {
			t.Parallel()
			checks := append([]rbacCheck{
				{"update", ns, true},
				{"patch", ns, true},
				{"delete", ns, true},
			}, shared...)
			assertRBAC(t, cluster, zoneMaintainerGroup, KyvernoTestNSName, checks)
			assertRBAC(t, cluster, zoneMaintainerGroup, "kube-system", outsideZone("kube-system"))
		})
		t.Run("managed-resource-guard", func(t *testing.T) {
			t.Parallel()
			testManagedResourceGuard(t, cluster)
		})
	})
}

// testManagedResourceGuard covers platform-apis-zone-managed-resource-deny. RBAC cannot stop a
// managed resource that reaches a Zone Namespace another way (ArgoCD, a self-granted Role), so the
// request is sent as the test's cluster admin: only the policy can reject it. Server-side dry run
// passes admission without creating anything.
func testManagedResourceGuard(t *testing.T, cluster *terrak8s.KubectlOptions) {
	manifest := `apiVersion: iam.aws.m.upbound.io/v1beta1
kind: RolePolicyAttachment
metadata:
  name: zone-rbac-escalation
  namespace: ` + KyvernoTestNSName + `
spec:
  forProvider:
    policyArn: arn:aws:iam::aws:policy/AdministratorAccess
    role: does-not-exist`
	out, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "create", "--dry-run=server", "-f", writeTempYAML(t, manifest))
	if err == nil {
		t.Fatalf("RolePolicyAttachment in a Zone Namespace was admitted: %s", out)
	}
	if !strings.Contains(out, "platform-apis-zone-managed-resource-deny") {
		t.Fatalf("RolePolicyAttachment was rejected, but not by platform-apis-zone-managed-resource-deny "+
			"(is the CRD installed?): %s", out)
	}
}

func assertRBAC(t *testing.T, cluster *terrak8s.KubectlOptions, group, namespace string, checks []rbacCheck) {
	t.Helper()
	for _, c := range checks {
		got := allowed(t, cluster, group, namespace, c.verb, c.resource)
		if got != c.want {
			t.Errorf("group %s: %s %s in %s allowed = %v, want %v", group, c.verb, c.resource, namespace, got, c.want)
		}
	}
}

// allowed creates a SubjectAccessReview for the group. Unlike kubectl auth can-i it sends the API group
// explicitly, so the answer does not depend on the CRD being installed in the cluster.
func allowed(t *testing.T, cluster *terrak8s.KubectlOptions, group, namespace, verb, resource string) bool {
	t.Helper()
	attrs := map[string]string{"namespace": namespace, "verb": verb}
	resource, attrs["group"], _ = strings.Cut(resource, ".")
	resource, sub, _ := strings.Cut(resource, "/")
	attrs["resource"] = resource
	if resource == "namespaces" {
		attrs["name"] = sub
	} else {
		attrs["subresource"] = sub
	}
	review, err := json.Marshal(map[string]any{
		"apiVersion": "authorization.k8s.io/v1",
		"kind":       "SubjectAccessReview",
		"spec": map[string]any{
			"user":               "zone-rbac-test",
			"groups":             []string{group},
			"resourceAttributes": attrs,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "create", "-f", writeTempYAML(t, string(review)),
		"-o", "jsonpath={.status.allowed}")
	if err != nil {
		t.Fatalf("SubjectAccessReview for %s %s: %v", verb, resource, err)
	}
	// status.allowed is omitted when false
	return strings.TrimSpace(out) == "true"
}
