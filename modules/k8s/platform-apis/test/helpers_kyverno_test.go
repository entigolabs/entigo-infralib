package test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type kyvernoNsData struct {
	Name, Zone, Enforce, Warn string
}

type kyvernoConfigMapData struct {
	Name, Namespace string
}

type kyvernoRoleData struct {
	Name, Namespace string
}

type kyvernoZoneData struct {
	Name       string
	Namespaces []string
}

type kyvernoArgoAppData struct {
	Name, Namespace, DestNamespace, Project string
	NamespaceLabels                         map[string]string
	NamespaceAnnotations                    map[string]string
}

type kyvernoKubeconfigData struct {
	CA, Server, Token string
}

type kyvernoRepositoryData struct {
	Name, Namespace string
}

type kyvernoPodData struct {
	Name, Namespace string
	NodeSelector    map[string]string
}

const (
	eksTokenPrefix        = "k8s-aws-v1."
	eksClusterIDHeader    = "x-k8s-aws-id"
	eksTokenExpirySeconds = "60"
	eksEmptyPayloadHash   = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// ensureKyvernoTestNamespace creates the shared kyverno-test namespace if it does not exist,
// then waits until the zone composition has reconciled the per-namespace Kyverno policies.
func ensureKyvernoTestNamespace(t *testing.T, cluster *terrak8s.KubectlOptions, z testZones) {
	t.Helper()
	applyFile(t, cluster, writeTempYAML(t, nsYAML(t, kyvernoNsData{
		Name: KyvernoTestNSName, Zone: z.A(), Enforce: "baseline", Warn: "baseline",
	})))
	waitNamespacePoliciesReady(t, cluster, z.A(), KyvernoTestNSName)
}

// waitNamespaceRoleBinding waits until the zone composition has created the per-namespace
// RoleBinding for the given role, confirming that role-mapping resources have landed.
func waitNamespaceRoleBinding(t *testing.T, cluster *terrak8s.KubectlOptions, namespace, role string) {
	t.Helper()
	nsOpts := terrak8s.NewKubectlOptions(cluster.ContextName, cluster.ConfigPath, namespace)
	waitResourceExists(t, nsOpts, "rolebindings", namespace+"-"+role, 30, 10*time.Second)
}

// waitNamespacePoliciesReady waits in parallel until the zone composition has created all
// per-namespace Kyverno policies, signaling that zone-managed resources for the namespace are ready.
func waitNamespacePoliciesReady(t *testing.T, cluster *terrak8s.KubectlOptions, zone, namespace string) {
	t.Helper()
	policies := []struct{ kind, name string }{
		{"mutatingpolicies.policies.kyverno.io", zone + "-" + namespace + "-add-nodeselector"},
		{"mutatingpolicies.policies.kyverno.io", zone + "-" + namespace + "-labels"},
		{"validatingpolicies.policies.kyverno.io", zone + "-" + namespace + "-validate-nodeselector"},
	}
	t.Run("wait-namespace-policies", func(t *testing.T) {
		for _, p := range policies {
			p := p
			t.Run(p.name, func(t *testing.T) {
				t.Parallel()
				waitResourceExists(t, cluster, p.kind, p.name, 30, 10*time.Second)
			})
		}
	})
}

// roleKubectlOptions builds a temporary kubeconfig that authenticates to the same EKS cluster
// using the provided AWS IAM credentials, enabling role-based webhook testing.
// The context name must be an EKS ARN: arn:aws:eks:<region>:<account>:cluster/<name>.
func roleKubectlOptions(t *testing.T, base *terrak8s.KubectlOptions, keyID, secret string) *terrak8s.KubectlOptions {
	t.Helper()

	parts := strings.SplitN(base.ContextName, ":", 6)
	region := parts[3]
	clusterName := parts[5]
	if idx := strings.LastIndex(clusterName, "/"); idx >= 0 {
		clusterName = clusterName[idx+1:]
	}

	ca, err := terrak8s.RunKubectlAndGetOutputE(t, base,
		"config", "view", "--raw", "--minify", "-o", "jsonpath={.clusters[0].cluster.certificate-authority-data}")
	require.NoError(t, err, "read cluster CA from kubeconfig")

	server, err := terrak8s.RunKubectlAndGetOutputE(t, base,
		"config", "view", "--minify", "-o", "jsonpath={.clusters[0].cluster.server}")
	require.NoError(t, err, "read cluster server from kubeconfig")

	kubeconfig := renderTemplate(t, "./templates/kyverno_kubeconfig.yaml", kyvernoKubeconfigData{
		CA: ca, Server: server, Token: eksToken(t, region, clusterName, keyID, secret),
	})

	f, ferr := os.CreateTemp("", "kubeconfig-role-*.yaml")
	require.NoError(t, ferr)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	_, err = f.WriteString(kubeconfig)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	return terrak8s.NewKubectlOptions("role-context", f.Name(), "")
}

func eksToken(t *testing.T, region, clusterName, keyID, secret string) string {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("https://sts.%s.amazonaws.com/?Action=GetCallerIdentity&Version=2011-06-15", region), nil)
	require.NoError(t, err, "build STS request")
	req.Header.Set(eksClusterIDHeader, clusterName)
	query := req.URL.Query()
	query.Set("X-Amz-Expires", eksTokenExpirySeconds)
	req.URL.RawQuery = query.Encode()

	creds := aws.Credentials{AccessKeyID: keyID, SecretAccessKey: secret}
	signed, _, err := v4.NewSigner().PresignHTTP(context.Background(), creds, req, eksEmptyPayloadHash, "sts", region, time.Now())
	require.NoError(t, err, "presign STS request")
	return fmt.Sprintf("%s%s", eksTokenPrefix, base64.RawURLEncoding.EncodeToString([]byte(signed)))
}

// kyvernoApply applies yamlStr so admission webhooks fire against real resources.
func kyvernoApply(t *testing.T, opts *terrak8s.KubectlOptions, yamlStr string) (string, error) {
	t.Helper()
	return terrak8s.RunKubectlAndGetOutputE(t, opts, "apply", "-f", writeTempYAML(t, yamlStr))
}

// writeTempYAML writes content to a temp file, schedules removal with t.Cleanup, and returns the path.
func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "kyverno-test-*.yaml")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	_, err = f.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

// renderTemplate loads templatePath, executes it with data, and returns the rendered string.
func renderTemplate(t *testing.T, templatePath string, data interface{}) string {
	t.Helper()
	raw, err := os.ReadFile(templatePath)
	require.NoError(t, err, "read template %s", templatePath)
	tmpl, err := template.New("").Parse(string(raw))
	require.NoError(t, err, "parse template %s", templatePath)
	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, data), "render template %s", templatePath)
	return buf.String()
}

// nsYAML renders the kyverno_namespace.yaml template.
// When Zone is empty the zone label is omitted (missing zone label test).
func nsYAML(t *testing.T, data kyvernoNsData) string {
	return renderTemplate(t, "./templates/kyverno_namespace.yaml", data)
}

// zoneYAML renders the kyverno_zone.yaml template.
func zoneYAML(t *testing.T, data kyvernoZoneData) string {
	return renderTemplate(t, "./templates/kyverno_zone.yaml", data)
}

// argoAppYAML renders the kyverno_argoapp.yaml template.
func argoAppYAML(t *testing.T, data kyvernoArgoAppData) string {
	return renderTemplate(t, "./templates/kyverno_argoapp.yaml", data)
}

func repositoryYAML(t *testing.T, data kyvernoRepositoryData) string {
	return renderTemplate(t, "./templates/kyverno_repository.yaml", data)
}

// configMapYAML renders the kyverno_configmap.yaml template.
func configMapYAML(t *testing.T, data kyvernoConfigMapData) string {
	return renderTemplate(t, "./templates/kyverno_configmap.yaml", data)
}

// roleYAML renders the kyverno_role.yaml template.
func roleYAML(t *testing.T, data kyvernoRoleData) string {
	return renderTemplate(t, "./templates/kyverno_role.yaml", data)
}

// podYAML renders the kyverno_pod.yaml template: a restricted-compliant pause pod whose
// nodeSelector is set verbatim, so the per-namespace nodeSelector policies see the given keys.
func podYAML(t *testing.T, data kyvernoPodData) string {
	return renderTemplate(t, "./templates/kyverno_pod.yaml", data)
}

// ── Assertion helpers ─────────────────────────────────────────────────────────

// assertKyvernoDenied asserts that the request was explicitly rejected by a Kyverno policy
func assertKyvernoDenied(t *testing.T, out string, err error) {
	t.Helper()
	combined := out
	if err != nil {
		combined += " " + err.Error()
	}
	require.Error(t, err, "expected Kyverno to deny the request; output: %s", combined)
	require.Contains(t, strings.ToLower(combined), "denied",
		"rejection should mention 'denied'; output: %s", combined)
}

// assertKyvernoAllowed asserts that Kyverno did not deny the request.
func assertKyvernoAllowed(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	require.NotContains(t, strings.ToLower(err.Error()), "denied",
		"expected Kyverno to allow the request but got a denial: %v", err)
}

// assertForbidden asserts that the operation was blocked by either RBAC ("forbidden") or
// Kyverno ("denied"). Used for role-based tests where the specific rejection layer may vary
// depending on the cluster's RBAC configuration.
func assertForbidden(t *testing.T, out string, err error) {
	t.Helper()
	combined := strings.ToLower(out)
	if err != nil {
		combined += " " + strings.ToLower(err.Error())
	}
	require.Error(t, err, "expected operation to be blocked (RBAC or Kyverno); output: %s", combined)
	require.True(t, strings.Contains(combined, "denied") || strings.Contains(combined, "forbidden"),
		"rejection should mention 'denied' or 'forbidden'; output: %s", combined)
}

// impersonatedKubectlOptions copies the base kubeconfig and makes its user impersonate
// user and groups, so admission sees an oidc: user as after a real login.
func impersonatedKubectlOptions(t *testing.T, base *terrak8s.KubectlOptions, user string, groups ...string) *terrak8s.KubectlOptions {
	t.Helper()
	raw, err := terrak8s.RunKubectlAndGetOutputE(t, base, "config", "view", "--raw", "--minify")
	require.NoError(t, err, "read kubeconfig")
	var cfg map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(raw), &cfg), "parse kubeconfig")
	users, ok := cfg["users"].([]any)
	require.True(t, ok && len(users) == 1, "minified kubeconfig has one user")
	entry, ok := users[0].(map[string]any)
	require.True(t, ok, "kubeconfig user entry")
	auth, ok := entry["user"].(map[string]any)
	require.True(t, ok, "kubeconfig user credentials")
	auth["as"] = user
	auth["as-groups"] = append(append([]string{}, groups...), "system:authenticated")
	out, err := yaml.Marshal(cfg)
	require.NoError(t, err, "write kubeconfig")
	return terrak8s.NewKubectlOptions(base.ContextName, writeTempYAML(t, string(out)), "")
}
