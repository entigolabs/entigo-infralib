package test

import (
	"encoding/json"
	"strings"
	"testing"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
)

const (
	zoneAContributorGroup = "zone-a-contributor"
	zoneMaintainerGroup   = "maintainer"
	zoneBNamespace        = "b1"
)

// rbacCheck resource is "resource[/subresource][.group]", or "namespaces/<name>".
type rbacCheck struct {
	verb, resource string
	want           bool
}

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
			{"create", "secretstores.external-secrets.io", true},
			{"create", "passwords.generators.external-secrets.io", true},
			{"create", "postgresqlinstances.database.entigo.com", true},
			{"create", "webapps.workload.entigo.com", true},
			{"create", "certificates.cert-manager.io", true},
			{"create", "issuers.cert-manager.io", true},
			{"list", "orders.acme.cert-manager.io", true},
			{"create", "servicemonitors.monitoring.coreos.com", true},
			// composed managed resources are readable for debugging, never writable
			{"get", "instances.rds.aws.m.upbound.io", true},
			{"list", "keys.kms.aws.m.upbound.io", true},
			{"watch", "securitygroups.ec2.aws.m.upbound.io", true},
			{"update", "instances.rds.aws.m.upbound.io", false},
			{"delete", "keys.kms.aws.m.upbound.io", false},
			{"create", "rolepolicyattachments.iam.aws.m.upbound.io", false},
			{"create", "roles.iam.aws.m.upbound.io", false},
			{"create", "buckets.s3.aws.m.upbound.io", false},
			{"create", "rolebindings.rbac.authorization.k8s.io", true},
			{"create", "roles.rbac.authorization.k8s.io", true},
			{"escalate", "roles.rbac.authorization.k8s.io", false},
			{"bind", "roles.rbac.authorization.k8s.io", false},
			{"delete", "networkpolicies.networking.k8s.io", true},
			{"create", "leases.coordination.k8s.io", true},
			{"list", "pods.metrics.k8s.io", true},
			{"create", "pods/attach", true},
			{"patch", "pods/ephemeralcontainers", true},
			{"update", "sidecars.networking.istio.io", false},
			{"create", "authorizationpolicies.security.istio.io", false},
			{"create", "policyexceptions.policies.kyverno.io", false},
			{"update", "resourcequotas", false},
		}

		outsideZone := func(namespace string) []rbacCheck {
			return []rbacCheck{
				{"update", "namespaces/" + namespace, false},
				{"patch", "namespaces/" + namespace, false},
				{"delete", "namespaces/" + namespace, false},
			}
		}
		otherZone := append([]rbacCheck{
			{"create", "deployments.apps", false},
			{"get", "instances.rds.aws.m.upbound.io", false},
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
		t.Run("rbac-escalation", func(t *testing.T) {
			t.Parallel()
			testRBACEscalation(t, cluster)
		})
		t.Run("managed-resource-guard", func(t *testing.T) {
			t.Parallel()
			testManagedResourceGuard(t, cluster)
		})
	})
}

// testRBACEscalation impersonates a contributor, so the API server's escalation check runs for real:
// a Role is limited to the permissions its creator holds, a SubjectAccessReview cannot show that.
func testRBACEscalation(t *testing.T, cluster *terrak8s.KubectlOptions) {
	as := []string{"--as=zone-rbac-test", "--as-group=" + zoneAContributorGroup, "--as-group=system:authenticated",
		"-n", KyvernoTestNSName, "--dry-run=server"}
	create := func(args ...string) (string, error) {
		return terrak8s.RunKubectlAndGetOutputE(t, cluster, append(append([]string{"create"}, args...), as...)...)
	}
	if out, err := create("role", "zone-rbac-subset", "--verb=get,list", "--resource=pods,configmaps"); err != nil {
		t.Errorf("contributor could not create a Role with a subset of its own permissions: %s", out)
	}
	// the Zone's own Role, a dry-run Role does not exist for a RoleBinding to refer to
	if out, err := create("rolebinding", "zone-rbac-subset", "--role="+KyvernoTestNSName+"-all",
		"--serviceaccount="+KyvernoTestNSName+":default"); err != nil {
		t.Errorf("contributor could not bind a Role it holds to a ServiceAccount: %s", out)
	}
	const escalation = "attempting to grant RBAC permissions not currently held"
	if out, err := create("role", "zone-rbac-escalation", "--verb=*", "--resource=*"); err == nil {
		t.Errorf("contributor created a Role with permissions it does not hold: %s", out)
	} else if !strings.Contains(out, escalation) {
		t.Errorf("Role was rejected, but not by the escalation check: %s", out)
	}
	if out, err := create("rolebinding", "zone-rbac-escalation", "--clusterrole=cluster-admin",
		"--serviceaccount="+KyvernoTestNSName+":default"); err == nil {
		t.Errorf("contributor bound cluster-admin in its Namespace: %s", out)
	} else if !strings.Contains(out, escalation) {
		t.Errorf("RoleBinding was rejected, but not by the escalation check: %s", out)
	}
}

// testManagedResourceGuard runs as cluster admin, so only the policy can reject the request.
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

// allowed uses a SubjectAccessReview, as kubectl auth can-i drops the API group of CRDs that are not installed.
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
