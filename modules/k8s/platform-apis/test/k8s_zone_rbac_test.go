package test

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-test/retry"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
)

// rbacCheck resource is "resource[/subresource][.group]", "namespaces/<name>" or "zones/<name>.tenancy.entigo.com".
type rbacCheck struct {
	verb, resource string
	want           bool
}

type zoneAuthz struct {
	Enabled                 bool
	Organization, Workspace string
}

type authzGrant struct {
	name, role, scope string
	namespaced        bool
}

const (
	zoneAContributorGroup = "zone-a-contributor"
	zoneMaintainerGroup   = "maintainer"
	zoneBNamespace        = "b1"
)

var authzGrants = []authzGrant{
	{name: "org-admin", role: "admin", scope: "organization", namespaced: true},
	{name: "e2e-workspace-owner", role: "workspace-owner", scope: "workspace", namespaced: true},
	{name: "e2e-iam-administrator", role: "iam-administrator", scope: "workspace"},
	{name: "e2e-contributor-a", role: "contributor", scope: "zone", namespaced: true},
	{name: "e2e-zone-maintainer-a", role: "zone-maintainer", scope: "zone", namespaced: true},
	{name: "e2e-zone-readonly-a", role: "zone-readonly", scope: "zone", namespaced: true},
	{name: "e2e-argocd-deployer-a", role: "argocd-deployer", scope: "zone", namespaced: true},
}

func readZoneAuthz(t *testing.T, cluster *terrak8s.KubectlOptions) zoneAuthz {
	t.Helper()
	out, err := terrak8s.RunKubectlAndGetOutputE(t, cluster, "get", "deploymentruntimeconfig", TenancyFunctionName,
		"-o", "jsonpath={.spec.deploymentTemplate.spec.template.spec.containers[0].env}")
	require.NoError(t, err, "read DeploymentRuntimeConfig %s", TenancyFunctionName)
	var a zoneAuthz
	if strings.TrimSpace(out) == "" {
		return a
	}
	var vars []struct{ Name, Value string }
	require.NoError(t, json.Unmarshal([]byte(out), &vars), "parse env of %s", TenancyFunctionName)
	for _, v := range vars {
		switch v.Name {
		case "AUTHZ_ENABLED":
			a.Enabled = v.Value == "true"
		case "AUTHZ_ORGANIZATION":
			a.Organization = v.Value
		case "AUTHZ_WORKSPACE":
			a.Workspace = v.Value
		}
	}
	return a
}

func (a zoneAuthz) zoneEntry(role, zone string) string {
	return fmt.Sprintf("%s@%s/%s/%s", role, a.Organization, a.Workspace, zone)
}

func (a zoneAuthz) workspaceEntry(role string) string {
	return fmt.Sprintf("%s@%s/%s", role, a.Organization, a.Workspace)
}

func (a zoneAuthz) organizationEntry(role string) string {
	return fmt.Sprintf("%s@%s", role, a.Organization)
}

func oidcGroup(entry string) string {
	return fmt.Sprintf("oidc:%s", entry)
}

func (g authzGrant) entry(a zoneAuthz, z testZones) string {
	switch g.scope {
	case "zone":
		return a.zoneEntry(g.role, z.A())
	case "organization":
		return a.organizationEntry(g.role)
	}
	return a.workspaceEntry(g.role)
}

func waitAuthzGrantsRendered(t *testing.T, cluster *terrak8s.KubectlOptions) {
	t.Helper()
	nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, KyvernoTestNSName)
	for _, g := range authzGrants {
		waitResourceExists(t, cluster, RoleGrantKind, g.name, 30, 10*time.Second)
		waitResourceExists(t, cluster, "clusterrolebindings", fmt.Sprintf("entigo-rg-%s-cluster", g.name), 30, 10*time.Second)
		if g.namespaced {
			waitResourceExists(t, nsOpts, "rolebindings", fmt.Sprintf("entigo-rg-%s", g.name), 30, 10*time.Second)
		}
	}
}

func testZoneRBAC(t *testing.T, cluster, argocd *terrak8s.KubectlOptions, z testZones) {
	if a := readZoneAuthz(t, cluster); a.Enabled {
		testZoneRBACAuthz(t, cluster, a, z)
		testZoneArgoCDAuthz(t, argocd, a, z)
		return
	}

	t.Run("zone-rbac", func(t *testing.T) {
		ensureKyvernoTestNamespace(t, cluster, z)
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
			// granted via namespaceRole.additionalRules; the ungranted verb and resource stay denied
			{"create", "widgets.example.entigo.com", true},
			{"get", "widgets.example.entigo.com", true},
			{"delete", "widgets.example.entigo.com", false},
			{"create", "gadgets.example.entigo.com", false},
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
	if resource == "namespaces" || resource == "zones" {
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

func testZoneRBACAuthz(t *testing.T, cluster *terrak8s.KubectlOptions, a zoneAuthz, z testZones) {
	t.Run("zone-rbac", func(t *testing.T) {
		ensureKyvernoTestNamespace(t, cluster, z)
		waitAuthzGrantsRendered(t, cluster)
		if t.Failed() {
			return
		}
		ns := KyvernoTestNSName
		zoneA := fmt.Sprintf("zones/%s.tenancy.entigo.com", z.A())
		zoneB := fmt.Sprintf("zones/%s.tenancy.entigo.com", z.B())
		type scoped struct {
			namespace string
			checks    []rbacCheck
		}
		want := map[string][]scoped{
			"admin": {
				{"kube-system", []rbacCheck{{"delete", "pods", true}}},
				{"", []rbacCheck{{"create", "clusterrolebindings.rbac.authorization.k8s.io", true}}},
			},
			"workspace-owner": {
				{"kube-system", []rbacCheck{{"create", "deployments.apps", true}}},
				{zoneBNamespace, []rbacCheck{{"create", "deployments.apps", true}}},
				{"", []rbacCheck{
					{"create", "clusterrolebindings.rbac.authorization.k8s.io", true},
					{"update", zoneB, true},
				}},
			},
			"iam-administrator": {
				{"", []rbacCheck{
					{"create", "rolegrants.authz.entigo.com", true},
					{"create", "groups.iam.meta.entigo.com", true},
				}},
				{ns, []rbacCheck{{"create", "deployments.apps", false}}},
				{"kube-system", []rbacCheck{{"get", "secrets", false}}},
			},
			"contributor": {
				{ns, []rbacCheck{
					{"create", "deployments.apps", true},
					{"get", "secrets", true},
					{"create", "rolebindings.rbac.authorization.k8s.io", true},
				}},
				{"", []rbacCheck{
					{"get", zoneA, true},
					{"update", zoneA, false},
					{"create", "rolegrants.authz.entigo.com", false},
				}},
				{zoneBNamespace, []rbacCheck{{"create", "deployments.apps", false}, {"get", "secrets", false}}},
				{"kube-system", []rbacCheck{{"get", "secrets", false}}},
			},
			"zone-maintainer": {
				{ns, []rbacCheck{{"create", "deployments.apps", true}}},
				{"", []rbacCheck{
					{"update", zoneA, true},
					{"delete", zoneA, true},
					{"update", zoneB, false},
					{"create", "zones.tenancy.entigo.com", false},
					{"create", "namespaces", true},
				}},
				{zoneBNamespace, []rbacCheck{{"create", "deployments.apps", false}}},
			},
			"zone-readonly": {
				{ns, []rbacCheck{
					{"get", "pods", true},
					{"list", "secrets", true},
					{"create", "deployments.apps", false},
					{"delete", "pods", false},
				}},
				{"", []rbacCheck{{"get", zoneA, true}, {"update", zoneA, false}}},
				{zoneBNamespace, []rbacCheck{{"get", "pods", false}}},
			},
			"argocd-deployer": {
				{ns, []rbacCheck{
					{"get", "applications.argoproj.io", true},
					{"sync", "applications.argoproj.io", true},
					{"create", "applications.argoproj.io", false},
					{"create", "deployments.apps", false},
				}},
				{zoneBNamespace, []rbacCheck{{"get", "applications.argoproj.io", false}}},
			},
		}
		for _, g := range authzGrants {
			g := g
			t.Run(g.role, func(t *testing.T) {
				t.Parallel()
				require.NotEmpty(t, want[g.role], "no checks for role %s", g.role)
				for _, s := range want[g.role] {
					assertRBAC(t, cluster, oidcGroup(g.entry(a, z)), s.namespace, s.checks)
				}
			})
		}
		t.Run("managed-resource-guard", func(t *testing.T) {
			t.Parallel()
			testManagedResourceGuard(t, cluster)
		})
	})
}

func testZoneArgoCDAuthz(t *testing.T, argocd *terrak8s.KubectlOptions, a zoneAuthz, z testZones) {
	t.Run("argocd-authz", func(t *testing.T) {
		t.Run("project-roles", func(t *testing.T) {
			t.Parallel()
			want := map[string]string{
				"admin":           a.organizationEntry("admin"),
				"workspace-owner": a.workspaceEntry("workspace-owner"),
				"contributor":     a.zoneEntry("contributor", z.A()),
				"zone-maintainer": a.zoneEntry("zone-maintainer", z.A()),
				"zone-readonly":   a.zoneEntry("zone-readonly", z.A()),
				"argocd-deployer": a.zoneEntry("argocd-deployer", z.A()),
			}
			cicd := []string{want["admin"], want["workspace-owner"], want["contributor"], want["zone-maintainer"]}
			_, err := retry.DoWithRetryE(t, fmt.Sprintf("AppProject %s roles from grants", z.A()), 30, 10*time.Second,
				func() (string, error) {
					out, err := terrak8s.RunKubectlAndGetOutputE(t, argocd, "get", AppProjectKind, z.A(), "-o", "jsonpath={.spec.roles}")
					if err != nil {
						return "", err
					}
					var roles []struct {
						Name   string
						Groups []string
					}
					if err := json.Unmarshal([]byte(out), &roles); err != nil {
						return "", err
					}
					groups := map[string][]string{}
					for _, r := range roles {
						groups[r.Name] = r.Groups
					}
					for role, entry := range want {
						if !slices.Contains(groups[role], entry) {
							return "", fmt.Errorf("role %s groups %v lack %s", role, groups[role], entry)
						}
					}
					for _, entry := range cicd {
						if !slices.Contains(groups["cicd"], entry) {
							return "", fmt.Errorf("role cicd groups %v lack %s", groups["cicd"], entry)
						}
					}
					return out, nil
				})
			require.NoError(t, err)
		})
		t.Run("rbac-cm", func(t *testing.T) {
			t.Parallel()
			for role, entry := range map[string]string{
				"admin":           a.organizationEntry("admin"),
				"workspace-owner": a.workspaceEntry("workspace-owner"),
			} {
				out, err := terrak8s.RunKubectlAndGetOutputE(t, argocd, "get", "configmap", ArgoCDRBACConfig,
					"-o", fmt.Sprintf("jsonpath={.data.policy\\.%s\\.csv}", role))
				require.NoError(t, err)
				require.Contains(t, out, fmt.Sprintf("g, %s, role:%s", entry, role), "%s policy.%s.csv", ArgoCDRBACConfig, role)
			}
		})
	})
}
