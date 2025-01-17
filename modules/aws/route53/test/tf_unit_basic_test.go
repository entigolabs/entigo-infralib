package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/stretchr/testify/assert"
)


func TestTerraformRoute53(t *testing.T) {
	t.Run("Biz", testTerraformRoute53Biz)
	t.Run("Pri", testTerraformRoute53Pri)
	t.Run("Min", testTerraformRoute53Min)
	t.Run("Ext", testTerraformRoute53Ext)
}

func testTerraformRoute53Biz(t *testing.T) {
        t.Parallel()
        outputs := aws.GetTFOutputs(t, "biz", "net")
	pub_zone_id := aws.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := aws.GetStringValue(t, outputs, "route53__pub_domain")
	pub_cert_arn := aws.GetStringValue(t, outputs, "route53__pub_cert_arn")
	int_zone_id := aws.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := aws.GetStringValue(t, outputs, "route53__int_domain")
	int_cert_arn := aws.GetStringValue(t, outputs, "route53__int_cert_arn")
	assert.NotEqual(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must not be equal")
	assert.Equal(t, "biz-net-route53-int.infralib.entigo.io", int_domain, "int_domain must be biz-net-route53-int.infralib.entigo.io")
	assert.Equal(t, "biz-net-route53.infralib.entigo.io", pub_domain, "pub_domain must be biz-net-route53.infralib.entigo.io")
	assert.NotEmpty(t, int_zone_id, "pub_domain was not returned")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	assert.NotEmpty(t, int_zone_id, "int_domain was not returned")
	assert.NotEmpty(t, int_cert_arn, "int_cert_arn was not returned")
}

func testTerraformRoute53Pri(t *testing.T) {
        t.Parallel()
        outputs := aws.GetTFOutputs(t, "pri", "net")
	pub_zone_id := aws.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := aws.GetStringValue(t, outputs, "route53__pub_domain")
	pub_cert_arn := aws.GetStringValue(t, outputs, "route53__pub_cert_arn")
	int_zone_id := aws.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := aws.GetStringValue(t, outputs, "route53__int_domain")
	assert.Equal(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, "pri-net-route53.infralib.entigo.io", int_domain, "int_domain must be pri-net-route53.infralib.entigo.io")
	assert.Equal(t, "pri-net-route53.infralib.entigo.io", pub_domain, "pub_domain must be pri-net-route53.infralib.entigo.io")
	assert.NotEmpty(t, int_zone_id, "int_zone_id was not returned")
	assert.NotEmpty(t, pub_zone_id, "pub_zone_id was not returned")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	
}

func testTerraformRoute53Min(t *testing.T) {
        //t.Parallel()
        outputs := aws.GetTFOutputs(t, "min", "net")
	pub_zone_id := aws.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := aws.GetStringValue(t, outputs, "route53__pub_domain")
	int_zone_id := aws.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := aws.GetStringValue(t, outputs, "route53__int_domain")
	assert.Empty(t, int_zone_id, "int_zone_id must be empty")
	assert.Empty(t, pub_zone_id, "pub_zone_id must be empty")
	assert.Equal(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, pub_domain, int_domain, "pub_domain and int_domain must be equal")
	assert.Equal(t, "infralib.entigo.io", int_domain, "int_domain must be infralib.entigo.io")
	assert.Equal(t, "infralib.entigo.io", pub_domain, "pub_domain must be infralib.entigo.io")

}

func testTerraformRoute53Ext(t *testing.T) {
        //t.Parallel()
        outputs := aws.GetTFOutputs(t, "ext", "net")
	pub_zone_id := aws.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := aws.GetStringValue(t, outputs, "route53__pub_domain")
	pub_cert_arn := aws.GetStringValue(t, outputs, "route53__pub_cert_arn")
	int_zone_id := aws.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := aws.GetStringValue(t, outputs, "route53__int_domain")
	assert.Equal(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, "mypub.infralib.entigo.io", int_domain, "int_domain must be mypub.infralib.entigo.io")
	assert.Equal(t, "mypub.infralib.entigo.io", pub_domain, "pub_domain must be mypub.infralib.entigo.io")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	assert.NotEmpty(t, int_zone_id, "int_zone_id was not returned")
	assert.NotEmpty(t, pub_zone_id, "pub_zone_id was not returned")

}

