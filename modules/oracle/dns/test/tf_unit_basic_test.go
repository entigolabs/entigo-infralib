package test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/oracle/oci-go-sdk/v65/certificatesmanagement"
	ocicommon "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oci.infralib.entigo.io, the parent_zone_id both test configurations delegate from.
const parentZoneId = "ocid1.dns-zone.oc1..aaaaaaaagynypxeneyf5an3eojldoicg6iyqj7od6nmd67kh5xjts5q2eq4q"

func TestTerraformDns(t *testing.T) {
	t.Run("Biz", testTerraformDnsBiz)
	t.Run("Pri", testTerraformDnsPri)
}

// biz.yaml has a public and a private domain, so int_* is the private one.
func testTerraformDnsBiz(t *testing.T) {
	t.Parallel()
	testTerraformDns(t, "biz", true)
}

// pri.yaml has only a public domain, so int_* falls back to it.
func testTerraformDnsPri(t *testing.T) {
	t.Parallel()
	testTerraformDns(t, "pri", false)
}

func testTerraformDns(t *testing.T, prefix string, withPrivate bool) {
	outputs := oracle.GetTFOutputs(t, prefix)
	// vpc and pca are in the net step, wherever this module is deployed.
	netOutputs := oracle.GetTFOutputsStep(t, prefix, "net")
	modulePrefix := fmt.Sprintf("%s-%s-dns", prefix, strings.ToLower(os.Getenv("STEP_NAME")))

	domains := map[string]interface{}{"public": modulePrefix + ".oci.infralib.entigo.io"}
	intKey := "public"
	if withPrivate {
		domains["private"] = modulePrefix + "-int.oci.infralib.entigo.io"
		intKey = "private"
	}

	assert.Equal(t, domains, tf.GetValue(t, outputs, "dns__domain_names"), "Wrong domain_names returned")
	assert.Equal(t, domains["public"], tf.GetStringValue(t, outputs, "dns__pub_domain"), "Wrong pub_domain returned")
	assert.Equal(t, domains[intKey], tf.GetStringValue(t, outputs, "dns__int_domain"), "Wrong int_domain returned")

	zoneIds := getMap(t, outputs, "dns__zone_ids")
	assert.ElementsMatch(t, keys(domains), keys(zoneIds), "zone_ids does not carry every domain")
	assert.Equal(t, zoneIds["public"], tf.GetStringValue(t, outputs, "dns__pub_zone_id"), "pub_zone_id is not the public zone")
	assert.Equal(t, zoneIds[intKey], tf.GetStringValue(t, outputs, "dns__int_zone_id"), "int_zone_id is not the %s zone", intKey)

	certificateIds := getMap(t, outputs, "dns__certificate_ocids")
	assert.ElementsMatch(t, keys(domains), keys(certificateIds), "certificate_ocids does not carry every domain")
	assert.Equal(t, certificateIds["public"], tf.GetStringValue(t, outputs, "dns__pub_cert_ocid"), "pub_cert_ocid is not the public certificate")
	assert.Equal(t, certificateIds[intKey], tf.GetStringValue(t, outputs, "dns__int_cert_ocid"), "int_cert_ocid is not the %s certificate", intKey)

	caId := tf.GetStringValue(t, netOutputs, "pca__certificate_authority_id")
	assert.Equal(t, caId, tf.GetStringValue(t, outputs, "dns__certificate_authority_id"), "certificate_authority_id is not the pca module's CA")

	// Every zone is created here, so every zone reports nameservers.
	nameservers := getMap(t, outputs, "dns__nameservers")
	assert.ElementsMatch(t, keys(domains), keys(nameservers), "nameservers does not carry every created zone")

	provider := configProvider()
	dnsClient, err := dns.NewDnsClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the DNS client")
	setRegion(&dnsClient)

	// The public zone is delegated from the parent with its own nameservers.
	publicZone := getZone(t, dnsClient, zoneIds["public"].(string))
	assert.Equal(t, domains["public"], *publicZone.Name, "Wrong public zone name")
	assert.Equal(t, dns.ScopeGlobal, publicZone.Scope, "Public zone is not GLOBAL")
	assert.Equal(t, dns.ZoneZoneTypePrimary, publicZone.ZoneType, "Public zone is not PRIMARY")
	publicNameservers := toStrings(nameservers["public"])
	zoneNameservers := []string{}
	for _, nameserver := range publicZone.Nameservers {
		zoneNameservers = append(zoneNameservers, *nameserver.Hostname)
	}
	assert.ElementsMatch(t, zoneNameservers, publicNameservers, "nameservers is not the public zone's")

	delegation, err := dnsClient.GetDomainRecords(context.Background(), dns.GetDomainRecordsRequest{
		ZoneNameOrId: ocicommon.String(parentZoneId),
		Domain:       ocicommon.String(domains["public"].(string)),
		Rtype:        ocicommon.String("NS"),
	})
	require.NoError(t, err, "Failed to read the NS delegation from the parent zone")
	delegatedNameservers := []string{}
	for _, record := range delegation.Items {
		delegatedNameservers = append(delegatedNameservers, *record.Rdata)
	}
	assert.ElementsMatch(t, publicNameservers, delegatedNameservers, "Parent zone does not delegate to the public zone's nameservers")

	// The private zone is in the view the vpc module's VCN resolves against.
	if withPrivate {
		privateZone := getZone(t, dnsClient, zoneIds["private"].(string))
		assert.Equal(t, domains["private"], *privateZone.Name, "Wrong private zone name")
		assert.Equal(t, dns.ScopePrivate, privateZone.Scope, "Private zone is not PRIVATE")
		assert.Equal(t, vcnDefaultViewId(t, provider, dnsClient, tf.GetStringValue(t, netOutputs, "vpc__vpc_id")), *privateZone.ViewId, "Private zone is not in the VCN's default view")
	}

	// Each domain gets a wildcard certificate from the pca module's CA.
	certificatesClient, err := certificatesmanagement.NewCertificatesManagementClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the Certificates Management client")
	setRegion(&certificatesClient)
	for key, domain := range domains {
		certificateId := certificateIds[key].(string)
		response, err := certificatesClient.GetCertificate(context.Background(), certificatesmanagement.GetCertificateRequest{CertificateId: &certificateId})
		require.NoError(t, err, "Failed to get the %s certificate", key)
		certificate := response.Certificate
		// name_salt appends a random suffix.
		assert.True(t, strings.HasPrefix(*certificate.Name, modulePrefix+"-"+key+"-"), "Wrong name for the %s certificate: %s", key, *certificate.Name)
		assert.Equal(t, certificatesmanagement.CertificateConfigTypeIssuedByInternalCa, certificate.ConfigType, "The %s certificate is not issued by an internal CA", key)
		assert.Equal(t, caId, *certificate.IssuerCertificateAuthorityId, "The %s certificate is not issued by the pca module's CA", key)
		assert.Equal(t, certificatesmanagement.CertificateLifecycleStateActive, certificate.LifecycleState, "The %s certificate is not active", key)
		assert.Equal(t, "*."+domain.(string), *certificate.Subject.CommonName, "Wrong common name for the %s certificate", key)
		sans := []string{}
		for _, san := range certificate.CurrentVersion.SubjectAlternativeNames {
			sans = append(sans, *san.Value)
		}
		assert.ElementsMatch(t, []string{"*." + domain.(string), domain.(string)}, sans, "Wrong SANs for the %s certificate", key)
	}
}

func getZone(t *testing.T, client dns.DnsClient, zoneId string) dns.Zone {
	response, err := client.GetZone(context.Background(), dns.GetZoneRequest{ZoneNameOrId: &zoneId})
	require.NoError(t, err, "Failed to get zone %s", zoneId)
	return response.Zone
}

// The default view of the resolver associated with the VCN.
func vcnDefaultViewId(t *testing.T, provider ocicommon.ConfigurationProvider, dnsClient dns.DnsClient, vcnId string) string {
	networkClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the Virtual Network client")
	setRegion(&networkClient)
	association, err := networkClient.GetVcnDnsResolverAssociation(context.Background(), core.GetVcnDnsResolverAssociationRequest{VcnId: &vcnId})
	require.NoError(t, err, "Failed to get the DNS resolver of VCN %s", vcnId)
	resolver, err := dnsClient.GetResolver(context.Background(), dns.GetResolverRequest{
		ResolverId: association.DnsResolverId,
		Scope:      dns.GetResolverScopePrivate,
	})
	require.NoError(t, err, "Failed to get resolver %s", *association.DnsResolverId)
	return *resolver.DefaultViewId
}

func configProvider() ocicommon.ConfigurationProvider {
	if configFile := os.Getenv("OCI_CONFIG_FILE"); configFile != "" {
		return ocicommon.CustomProfileConfigProvider(configFile, "DEFAULT")
	}
	return ocicommon.DefaultConfigProvider()
}

func setRegion(client interface{ SetRegion(string) }) {
	if region := os.Getenv("OCI_REGION"); region != "" {
		client.SetRegion(region)
	}
}

func getMap(t *testing.T, outputs map[string]interface{}, key string) map[string]interface{} {
	value, ok := tf.GetValue(t, outputs, key).(map[string]interface{})
	require.True(t, ok, "%s is not a map", key)
	return value
}

func keys(m map[string]interface{}) []string {
	result := []string{}
	for key := range m {
		result = append(result, key)
	}
	return result
}

func toStrings(value interface{}) []string {
	result := []string{}
	for _, item := range value.([]interface{}) {
		result = append(result, item.(string))
	}
	return result
}
