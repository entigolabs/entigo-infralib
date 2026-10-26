package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/entigolabs/entigo-infralib-test/aws"
	"github.com/entigolabs/entigo-infralib-test/env"
	"github.com/entigolabs/entigo-infralib-test/tf"
)

func TestEc2(t *testing.T) {
	env.RunEach(t, map[string]env.TestFunc{
		"aws_biz": testEc2Biz,
		"aws_pri": testEc2Pri,
	})
}

// biz runs the instance in a private subnet without a public ip.
func testEc2Biz(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)

	privateDns := outputs.String(t, "ec2__private_dns")
	assert.NotEmpty(t, privateDns, "private_dns must not be empty")

	assert.False(t, outputs.HasKeyWithPrefix("ec2__public_ip"), "Must not contain any ec2__public_ip outputs.")
}

func testEc2Pri(t *testing.T, e *env.Environment) {
	outputs := tf.Get(t, e)
	privateDns := outputs.String(t, "ec2__private_dns")
	publicIp := outputs.String(t, "ec2__public_ip")
	assert.NotEmpty(t, privateDns, "private_dns must not be empty")
	assert.NotEmpty(t, publicIp, "public_ip must not be empty")
}
