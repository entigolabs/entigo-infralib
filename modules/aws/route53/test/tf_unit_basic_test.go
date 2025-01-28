package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
	"strings"
	"fmt"
	"os"
)


func TestTerraformRoute53(t *testing.T) {
	t.Run("Biz", testTerraformRoute53Biz)
	t.Run("Pri", testTerraformRoute53Pri)
	t.Run("Min", testTerraformRoute53Min)
	t.Run("Ext", testTerraformRoute53Ext)
}

func testTerraformRoute53Biz(t *testing.T) {
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
	assert.NotEmpty(t, int_zone_id, "pub_domain was not returned")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	assert.NotEmpty(t, int_zone_id, "int_domain was not returned")
	assert.NotEmpty(t, int_cert_arn, "int_cert_arn was not returned")
}

func testTerraformRoute53Pri(t *testing.T) {
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
	assert.NotEmpty(t, int_zone_id, "int_zone_id was not returned")
	assert.NotEmpty(t, pub_zone_id, "pub_zone_id was not returned")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	
}

func testTerraformRoute53Min(t *testing.T) {
        //t.Parallel()
        outputs := aws.GetTFOutputs(t, "min")
	pub_zone_id := tf.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := tf.GetStringValue(t, outputs, "route53__pub_domain")
	int_zone_id := tf.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := tf.GetStringValue(t, outputs, "route53__int_domain")
	assert.Empty(t, int_zone_id, "int_zone_id must be empty")
	assert.Empty(t, pub_zone_id, "pub_zone_id must be empty")
	assert.Equal(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, pub_domain, int_domain, "pub_domain and int_domain must be equal")
	assert.Equal(t, "infralib.entigo.io", int_domain, "int_domain must be infralib.entigo.io")
	assert.Equal(t, "infralib.entigo.io", pub_domain, "pub_domain must be infralib.entigo.io")

}

func testTerraformRoute53Ext(t *testing.T) {
        //t.Parallel()
        outputs := aws.GetTFOutputs(t, "ext")
	pub_zone_id := tf.GetStringValue(t, outputs, "route53__pub_zone_id")
	pub_domain  := tf.GetStringValue(t, outputs, "route53__pub_domain")
	pub_cert_arn := tf.GetStringValue(t, outputs, "route53__pub_cert_arn")
	int_zone_id := tf.GetStringValue(t, outputs, "route53__int_zone_id")
	int_domain := tf.GetStringValue(t, outputs, "route53__int_domain")
	assert.Equal(t, int_zone_id, pub_zone_id, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, "mypub.infralib.entigo.io", int_domain, "int_domain must be mypub.infralib.entigo.io")
	assert.Equal(t, "mypub.infralib.entigo.io", pub_domain, "pub_domain must be mypub.infralib.entigo.io")
	assert.NotEmpty(t, pub_cert_arn, "pub_cert_arn was not returned")
	assert.NotEmpty(t, int_zone_id, "int_zone_id was not returned")
	assert.NotEmpty(t, pub_zone_id, "pub_zone_id was not returned")

}

