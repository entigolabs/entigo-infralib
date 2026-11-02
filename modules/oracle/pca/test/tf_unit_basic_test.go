package test

import (
	"context"
	"strings"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/certificatesmanagement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/oracle"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

// Both test configurations adopt an existing CA named <prefix>-net-pca-ca instead of creating one.
func TestPca(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		caName := e.Prefix + "-net-pca-ca"

		caId := outputs.String(t, "pca__certificate_authority_id")
		require.True(t, strings.HasPrefix(caId, "ocid1.certificateauthority."), "certificate_authority_id is not a CA OCID: %q", caId)

		assert.Equal(t, caName, outputs.String(t, "pca__certificate_authority_name"), "Wrong value for certificate_authority_name returned")

		bundleCommand := outputs.String(t, "pca__certificate_authority_bundle_command")
		assert.Contains(t, bundleCommand, caId, "certificate_authority_bundle_command does not name the adopted CA")

		// The name output repeats ca_name, so the adopted CA itself is read.
		ca := getCertificateAuthority(t, e, caId)
		assert.Equal(t, caName, *ca.Name, "certificate_authority_id is not the CA named by ca_name")
		assert.Equal(t, certificatesmanagement.CertificateAuthorityLifecycleStateActive, ca.LifecycleState, "Adopted CA is not active")
	})
}

func getCertificateAuthority(t *testing.T, e *env.Environment, caId string) certificatesmanagement.CertificateAuthority {
	client, err := certificatesmanagement.NewCertificatesManagementClientWithConfigurationProvider(oracle.ConfigProvider())
	require.NoError(t, err, "Failed to create the Certificates Management client")
	client.SetRegion(e.MustRegion(t))
	response, err := client.GetCertificateAuthority(context.Background(), certificatesmanagement.GetCertificateAuthorityRequest{CertificateAuthorityId: &caId})
	require.NoError(t, err, "Failed to get certificate authority %s", caId)
	return response.CertificateAuthority
}
