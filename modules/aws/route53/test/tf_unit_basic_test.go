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

// The module is part of no environment (aws-v2/route53 replaced it), so this
// test skips; the expectations per prefix are kept for when it is added back.
func TestRoute53(t *testing.T) {
	env.Run(t, func(t *testing.T, e *env.Environment) {
		outputs := tf.Get(t, e)
		stepName := strings.TrimSpace(strings.ToLower(env.ModulePlacement(t, e).Step.Name))

		pubZoneId := outputs.String(t, "route53__pub_zone_id")
		pubDomain := outputs.String(t, "route53__pub_domain")
		intZoneId := outputs.String(t, "route53__int_zone_id")
		intDomain := outputs.String(t, "route53__int_domain")

		switch e.Prefix {
		case "biz":
			// biz creates a separate private zone with its own certificate.
			pubCertArn := outputs.String(t, "route53__pub_cert_arn")
			intCertArn := outputs.String(t, "route53__int_cert_arn")
			assert.NotEqual(t, intZoneId, pubZoneId, "int_zone_id and pub_zone_id must not be equal")
			assert.Equal(t, fmt.Sprintf("biz-%s-route53-int.infralib.entigo.io", stepName), intDomain, "int_domain value is wrong")
			assert.Equal(t, fmt.Sprintf("biz-%s-route53.infralib.entigo.io", stepName), pubDomain, "pub_domain value is wrong")
			assert.NotEmpty(t, intZoneId, "pub_domain was not returned")
			assert.NotEmpty(t, pubCertArn, "pub_cert_arn was not returned")
			assert.NotEmpty(t, intZoneId, "int_domain was not returned")
			assert.NotEmpty(t, intCertArn, "int_cert_arn was not returned")
		case "pri":
			// pri uses the public zone for internal names as well.
			pubCertArn := outputs.String(t, "route53__pub_cert_arn")
			assert.Equal(t, intZoneId, pubZoneId, "int_zone_id and pub_zone_id must be equal")
			assert.Equal(t, fmt.Sprintf("pri-%s-route53.infralib.entigo.io", stepName), intDomain, "int_domain value is wrong")
			assert.Equal(t, fmt.Sprintf("pri-%s-route53.infralib.entigo.io", stepName), pubDomain, "pub_domain value is wrong")
			assert.NotEmpty(t, intZoneId, "int_zone_id was not returned")
			assert.NotEmpty(t, pubZoneId, "pub_zone_id was not returned")
			assert.NotEmpty(t, pubCertArn, "pub_cert_arn was not returned")
		case "min":
			// min creates no zones and only passes the parent domain through.
			assert.Empty(t, intZoneId, "int_zone_id must be empty")
			assert.Empty(t, pubZoneId, "pub_zone_id must be empty")
			assert.Equal(t, intZoneId, pubZoneId, "int_zone_id and pub_zone_id must be equal")
			assert.Equal(t, pubDomain, intDomain, "pub_domain and int_domain must be equal")
			assert.Equal(t, "infralib.entigo.io", intDomain, "int_domain must be infralib.entigo.io")
			assert.Equal(t, "infralib.entigo.io", pubDomain, "pub_domain must be infralib.entigo.io")
		case "ext":
			// ext uses an existing zone for both the public and the internal domain.
			pubCertArn := outputs.String(t, "route53__pub_cert_arn")
			assert.Equal(t, intZoneId, pubZoneId, "int_zone_id and pub_zone_id must be equal")
			assert.Equal(t, "mypub.infralib.entigo.io", intDomain, "int_domain must be mypub.infralib.entigo.io")
			assert.Equal(t, "mypub.infralib.entigo.io", pubDomain, "pub_domain must be mypub.infralib.entigo.io")
			assert.NotEmpty(t, pubCertArn, "pub_cert_arn was not returned")
			assert.NotEmpty(t, intZoneId, "int_zone_id was not returned")
			assert.NotEmpty(t, pubZoneId, "pub_zone_id was not returned")
		default:
			t.Fatalf("no expectations for environment %s", e.Name)
		}
	})
}
