package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestRoute53(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"aws_biz": testRoute53Biz,
		"aws_pri": testRoute53Pri,
		"aws_min": testRoute53Min,
		"aws_ext": testRoute53Ext,
	})
}

func moduleStepName(t *testing.T, e *env.Environment) string {
	return strings.TrimSpace(strings.ToLower(env.ModulePlacement(t, e).Step.Name))
}

// biz creates a separate private zone with its own certificate.
func testRoute53Biz(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	stepName := moduleStepName(t, e)

	pubZoneId := outputs.String(t, "route53__pub_zone_id")
	pubDomain := outputs.String(t, "route53__pub_domain")
	pubCertArn := outputs.String(t, "route53__pub_cert_arn")
	intZoneId := outputs.String(t, "route53__int_zone_id")
	intDomain := outputs.String(t, "route53__int_domain")
	intCertArn := outputs.String(t, "route53__int_cert_arn")
	assert.NotEqual(t, intZoneId, pubZoneId, "int_zone_id and pub_zone_id must not be equal")
	assert.Equal(t, fmt.Sprintf("biz-%s-route53-int.infralib.entigo.io", stepName), intDomain, "int_domain value is wrong")
	assert.Equal(t, fmt.Sprintf("biz-%s-route53.infralib.entigo.io", stepName), pubDomain, "pub_domain value is wrong")
	assert.NotEmpty(t, intZoneId, "pub_domain was not returned")
	assert.NotEmpty(t, pubCertArn, "pub_cert_arn was not returned")
	assert.NotEmpty(t, intZoneId, "int_domain was not returned")
	assert.NotEmpty(t, intCertArn, "int_cert_arn was not returned")
}

// pri uses the public zone for internal names as well.
func testRoute53Pri(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	stepName := moduleStepName(t, e)

	pubZoneId := outputs.String(t, "route53__pub_zone_id")
	pubDomain := outputs.String(t, "route53__pub_domain")
	pubCertArn := outputs.String(t, "route53__pub_cert_arn")
	intZoneId := outputs.String(t, "route53__int_zone_id")
	intDomain := outputs.String(t, "route53__int_domain")
	assert.Equal(t, intZoneId, pubZoneId, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, fmt.Sprintf("pri-%s-route53.infralib.entigo.io", stepName), intDomain, "int_domain value is wrong")
	assert.Equal(t, fmt.Sprintf("pri-%s-route53.infralib.entigo.io", stepName), pubDomain, "pub_domain value is wrong")
	assert.NotEmpty(t, intZoneId, "int_zone_id was not returned")
	assert.NotEmpty(t, pubZoneId, "pub_zone_id was not returned")
	assert.NotEmpty(t, pubCertArn, "pub_cert_arn was not returned")
}

// min reuses the existing parent zone for both domains.
func testRoute53Min(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	pubZoneId := outputs.String(t, "route53__pub_zone_id")
	pubDomain := outputs.String(t, "route53__pub_domain")
	intZoneId := outputs.String(t, "route53__int_zone_id")
	intDomain := outputs.String(t, "route53__int_domain")
	assert.NotEmpty(t, intZoneId, "int_zone_id must be empty")
	assert.NotEmpty(t, pubZoneId, "pub_zone_id must be empty")
	assert.Equal(t, intZoneId, pubZoneId, "int_zone_id and pub_zone_id must be equal")
	assert.Equal(t, pubDomain, intDomain, "pub_domain and int_domain must be equal")
	assert.Equal(t, "infralib.entigo.io", intDomain, "int_domain must be infralib.entigo.io")
	assert.Equal(t, "infralib.entigo.io", pubDomain, "pub_domain must be infralib.entigo.io")
}

// ext creates the public zone and a private zone with a custom name.
func testRoute53Ext(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	stepName := moduleStepName(t, e)
	pubZoneId := outputs.String(t, "route53__pub_zone_id")
	pubDomain := outputs.String(t, "route53__pub_domain")
	pubCertArn := outputs.String(t, "route53__pub_cert_arn")
	intZoneId := outputs.String(t, "route53__int_zone_id")
	intDomain := outputs.String(t, "route53__int_domain")
	assert.NotEqual(t, intZoneId, pubZoneId, "int_zone_id and pub_zone_id must not be equal")
	assert.Equal(t, fmt.Sprintf("ext-%s-route53-private.infralib.entigo.io", stepName), intDomain, "int_domain value is wrong")
	assert.Equal(t, fmt.Sprintf("ext-%s-route53.infralib.entigo.io", stepName), pubDomain, "pub_domain value is wrong")
	assert.NotEmpty(t, pubCertArn, "pub_cert_arn was not returned")
	assert.NotEmpty(t, intZoneId, "int_zone_id was not returned")
	assert.NotEmpty(t, pubZoneId, "pub_zone_id was not returned")
}
