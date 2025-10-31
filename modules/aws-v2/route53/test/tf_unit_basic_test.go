package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)


func TestTerraform	(t *testing.T) {
	t.Run("Biz", testTerraformRoute53v2Biz)
	t.Run("Pri", testTerraformRoute53v2Pri)
	t.Run("Min", testTerraformRoute53v2Min)
	t.Run("Ext", testTerraformRoute53v2Ext)
}

func testTerraformRoute53v2Biz(t *testing.T) {
        t.Parallel()
        outputs := aws.GetTFOutputs(t, "biz")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	
	pub_zone_id := tf.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := tf.GetStringValue(t, outputs, "route53__pub_domain")
	pub_cert_arn := tf.GetStringValue(t, outputs, "route53__pub_cert_arn")
	int_zone_id := tf.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := tf.GetStringValue(t, outputs, "route53__int_domain")
	int_cert_arn := tf.GetStringValue(t, outputs, "route53__int_cert_arn")
	assert.NotEqual(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must not be equal")
	assert.Equal(t, fmt.Sprintf("biz-%s-route53-int.infralib.entigo.io", stepName), int_domain, "int_domain value is wrong")
	assert.Equal(t, fmt.Sprintf("biz-%s-route53.infralib.entigo.io", stepName), pub_domain, "pub_domain value is wrong")
	//assert.Equal(t, "biz-route53-int.infralib.entigo.io", int_domain, "int_domain value is wrong")
	//assert.Equal(t, "biz-route53.infralib.entigo.io", pub_domain, "pub_domain value is wrong")
	assert.NotEmpty(t, int_zone_id, "pub_domain was not returned")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	assert.NotEmpty(t, int_zone_id, "int_domain was not returned")
	assert.NotEmpty(t, int_cert_arn, "int_cert_arn was not returned")
}

func testTerraformRoute53v2Pri(t *testing.T) {
        t.Parallel()
        outputs := aws.GetTFOutputs(t, "pri")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	
	pub_zone_id := tf.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := tf.GetStringValue(t, outputs, "route53__pub_domain")
	pub_cert_arn := tf.GetStringValue(t, outputs, "route53__pub_cert_arn")
	int_zone_id := tf.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := tf.GetStringValue(t, outputs, "route53__int_domain")
	assert.Equal(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, fmt.Sprintf("pri-%s-route53.infralib.entigo.io", stepName), int_domain, "int_domain value is wrong")
	assert.Equal(t, fmt.Sprintf("pri-%s-route53.infralib.entigo.io", stepName), pub_domain, "pub_domain value is wrong")
	//assert.Equal(t, "pri-route53.infralib.entigo.io", int_domain, "int_domain value is wrong")
	//assert.Equal(t, "pri-route53.infralib.entigo.io", pub_domain, "pub_domain value is wrong")
	assert.NotEmpty(t, int_zone_id, "int_zone_id was not returned")
	assert.NotEmpty(t, pub_zone_id, "pub_zone_id was not returned")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	
}

func testTerraformRoute53v2Min(t *testing.T) {
        //t.Parallel()
        outputs := aws.GetTFOutputs(t, "min")
	pub_zone_id := tf.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := tf.GetStringValue(t, outputs, "route53__pub_domain")
	int_zone_id := tf.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := tf.GetStringValue(t, outputs, "route53__int_domain")
	assert.NotEmpty(t, int_zone_id, "int_zone_id must be empty")
	assert.NotEmpty(t, pub_zone_id, "pub_zone_id must be empty")
	assert.Equal(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, pub_domain, int_domain, "pub_domain and int_domain must be equal")
	assert.Equal(t, "infralib.entigo.io", int_domain, "int_domain must be infralib.entigo.io")
	assert.Equal(t, "infralib.entigo.io", pub_domain, "pub_domain must be infralib.entigo.io")

}

func testTerraformRoute53v2Ext(t *testing.T) {
        //t.Parallel()
        outputs := aws.GetTFOutputs(t, "ext")
	stepName := strings.TrimSpace(strings.ToLower(os.Getenv("STEP_NAME")))
	pub_zone_id := tf.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := tf.GetStringValue(t, outputs, "route53__pub_domain")
	pub_cert_arn := tf.GetStringValue(t, outputs, "route53__pub_cert_arn")
	int_zone_id := tf.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := tf.GetStringValue(t, outputs, "route53__int_domain")
	assert.NotEqual(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must not be equal")
	assert.Equal(t, fmt.Sprintf("ext-%s-route53-private.infralib.entigo.io", stepName), int_domain, "int_domain value is wrong")
	assert.Equal(t, fmt.Sprintf("ext-%s-route53.infralib.entigo.io", stepName), pub_domain, "pub_domain value is wrong")
	//assert.Equal(t, "ext-route53-private.infralib.entigo.io", int_domain, "int_domain must be ext-route53-private.infralib.entigo.io")
	//assert.Equal(t, "ext-route53.infralib.entigo.io", pub_domain, "pub_domain must be ext-route53.infralib.entigo.io")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	assert.NotEmpty(t, int_zone_id, "int_zone_id was not returned")
	assert.NotEmpty(t, pub_zone_id, "pub_zone_id was not returned")

}

