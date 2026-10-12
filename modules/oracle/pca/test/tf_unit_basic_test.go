package test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/oracle"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/oracle/oci-go-sdk/v65/certificatesmanagement"
	ocicommon "github.com/oracle/oci-go-sdk/v65/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformPca(t *testing.T) {
	t.Run("Biz", testTerraformPcaBiz)
	t.Run("Pri", testTerraformPcaPri)
}

func testTerraformPcaBiz(t *testing.T) {
	t.Parallel()
	testTerraformPca(t, "biz", "biz-net-pca-root-ca")
}

func testTerraformPcaPri(t *testing.T) {
	t.Parallel()
	testTerraformPca(t, "pri", "pri-net-pca-root-ca")
}

// Both test configurations adopt an existing CA by ca_name instead of creating one.
func testTerraformPca(t *testing.T, prefix string, caName string) {
	outputs := oracle.GetTFOutputs(t, prefix)

	caId := tf.GetStringValue(t, outputs, "pca__certificate_authority_id")
	require.True(t, strings.HasPrefix(caId, "ocid1.certificateauthority."), "certificate_authority_id is not a CA OCID: %q", caId)

	assert.Equal(t, caName, tf.GetStringValue(t, outputs, "pca__certificate_authority_name"), "Wrong value for certificate_authority_name returned")

	bundleCommand := tf.GetStringValue(t, outputs, "pca__certificate_authority_bundle_command")
	assert.Contains(t, bundleCommand, caId, "certificate_authority_bundle_command does not name the adopted CA")

	// The name output repeats ca_name, so the adopted CA itself is read.
	ca := getCertificateAuthority(t, caId)
	assert.Equal(t, caName, *ca.Name, "certificate_authority_id is not the CA named by ca_name")
	assert.Equal(t, certificatesmanagement.CertificateAuthorityLifecycleStateActive, ca.LifecycleState, "Adopted CA is not active")
}

func getCertificateAuthority(t *testing.T, caId string) certificatesmanagement.CertificateAuthority {
	provider := ocicommon.DefaultConfigProvider()
	if configFile := os.Getenv("OCI_CONFIG_FILE"); configFile != "" {
		provider = ocicommon.CustomProfileConfigProvider(configFile, "DEFAULT")
	}
	client, err := certificatesmanagement.NewCertificatesManagementClientWithConfigurationProvider(provider)
	require.NoError(t, err, "Failed to create the Certificates Management client")
	if region := os.Getenv("OCI_REGION"); region != "" {
		client.SetRegion(region)
	}
	response, err := client.GetCertificateAuthority(context.Background(), certificatesmanagement.GetCertificateAuthorityRequest{CertificateAuthorityId: &caId})
	require.NoError(t, err, "Failed to get certificate authority %s", caId)
	return response.CertificateAuthority
}
