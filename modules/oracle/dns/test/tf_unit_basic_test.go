package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/certificatesmanagement"
	ocicommon "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/oracle"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestDns(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"oracle_biz": testDnsBiz,
		"oracle_pri": testDnsPri,
	})
}

// oracle_biz.yaml has two private domains with private as the default, so pub_* fall back to it.
func testDnsBiz(t *testing.T, e *env.Environment) {
	testDns(t, e, true)
}

// oracle_pri.yaml has a single private domain, so it is both the public and the private default.
func testDnsPri(t *testing.T, e *env.Environment) {
	testDns(t, e, false)
}

func testDns(t *testing.T, e *env.Environment, withExtra bool) {
	outputs := tf.Get(t, e)
	// vpc and pca are in the net step, wherever this module is deployed.
	netOutputs := tf.GetStep(t, e, "net")
	modulePrefix := fmt.Sprintf("%s-%s-dns", e.Prefix, env.ModulePlacement(t, e).Step.Name)

	domains := map[string]interface{}{"private": modulePrefix + "-int.oci.infralib.entigo.io"}
	if withExtra {
		domains["extra"] = modulePrefix + "-extra.oci.infralib.entigo.io"
	}

	assert.Equal(t, domains, outputs.Value(t, "dns__domain_names"), "Wrong domain_names returned")
	assert.Equal(t, domains["private"], outputs.String(t, "dns__pub_domain"), "Wrong pub_domain returned")
	assert.Equal(t, domains["private"], outputs.String(t, "dns__int_domain"), "Wrong int_domain returned")
	assert.Equal(t, "", outputs.String(t, "dns__public_domain"), "public_domain is set without a public zone")
	assert.Equal(t, domains["private"], outputs.String(t, "dns__private_domain"), "Wrong private_domain returned")

	zoneIds := outputs.Map(t, "dns__zone_ids")
	assert.ElementsMatch(t, keys(domains), keys(zoneIds), "zone_ids does not carry every domain")
	assert.Equal(t, zoneIds["private"], outputs.String(t, "dns__pub_zone_id"), "pub_zone_id is not the default private zone")
	assert.Equal(t, zoneIds["private"], outputs.String(t, "dns__int_zone_id"), "int_zone_id is not the default private zone")

	certificateIds := outputs.Map(t, "dns__certificate_ocids")
	assert.ElementsMatch(t, keys(domains), keys(certificateIds), "certificate_ocids does not carry every domain")
	assert.Equal(t, certificateIds["private"], outputs.String(t, "dns__pub_cert_ocid"), "pub_cert_ocid is not the default private certificate")
	assert.Equal(t, certificateIds["private"], outputs.String(t, "dns__int_cert_ocid"), "int_cert_ocid is not the default private certificate")

	caId := netOutputs.String(t, "pca__certificate_authority_id")
	assert.Equal(t, caId, outputs.String(t, "dns__certificate_authority_id"), "certificate_authority_id is not the pca module's CA")

	// Every zone is created here, so every zone reports nameservers.
	nameservers := outputs.Map(t, "dns__nameservers")
	assert.ElementsMatch(t, keys(domains), keys(nameservers), "nameservers does not carry every created zone")

	provider := oracle.ConfigProvider()
	region := e.MustRegion(t)
	dnsClient, err := dns.NewDnsClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the DNS client")
	dnsClient.SetRegion(region)

	// Every zone is in the view the vpc module's VCN resolves against.
	viewId := vcnDefaultViewId(t, provider, region, dnsClient, netOutputs.String(t, "vpc__vpc_id"))
	for key, domain := range domains {
		zone := getZone(t, dnsClient, zoneIds[key].(string))
		assert.Equal(t, domain, *zone.Name, "Wrong name for the %s zone", key)
		assert.Equal(t, dns.ScopePrivate, zone.Scope, "The %s zone is not PRIVATE", key)
		assert.Equal(t, viewId, *zone.ViewId, "The %s zone is not in the VCN's default view", key)
	}

	// Each domain gets a wildcard certificate from the pca module's CA.
	certificatesClient, err := certificatesmanagement.NewCertificatesManagementClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the Certificates Management client")
	certificatesClient.SetRegion(region)
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
func vcnDefaultViewId(t *testing.T, provider ocicommon.ConfigurationProvider, region string, dnsClient dns.DnsClient, vcnId string) string {
	networkClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the Virtual Network client")
	networkClient.SetRegion(region)
	association, err := networkClient.GetVcnDnsResolverAssociation(context.Background(), core.GetVcnDnsResolverAssociationRequest{VcnId: &vcnId})
	require.NoError(t, err, "Failed to get the DNS resolver of VCN %s", vcnId)
	resolver, err := dnsClient.GetResolver(context.Background(), dns.GetResolverRequest{
		ResolverId: association.DnsResolverId,
		Scope:      dns.GetResolverScopePrivate,
	})
	require.NoError(t, err, "Failed to get resolver %s", *association.DnsResolverId)
	return *resolver.DefaultViewId
}

func keys(m map[string]interface{}) []string {
	result := []string{}
	for key := range m {
		result = append(result, key)
	}
	return result
}
