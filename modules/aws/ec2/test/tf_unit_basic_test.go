package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/stretchr/testify/assert"
)



func TestTerraformEc2(t *testing.T) {
	t.Run("Biz", testTerraformEc2Biz)
	t.Run("Pri", testTerraformEc2Pri)
}

func testTerraformEc2Biz(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz", "infra")
	
	private_dns := aws.GetStringValue(t, outputs, "ec2__private_dns")
	assert.NotEmpty(t, private_dns, "private_dns must not be empty")
	
	assert.False(t, aws.HasKeyWithPrefix(t, outputs, "ec2__public_ip"), "Must not contain any ec2__public_ip outputs.")
}

func testTerraformEc2Pri(t *testing.T) {
	t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri", "infra")
	private_dns := aws.GetStringValue(t, outputs, "ec2__private_dns")
	public_ip := aws.GetStringValue(t, outputs, "ec2__public_ip")
	assert.NotEmpty(t, private_dns, "private_dns must not be empty")
	assert.NotEmpty(t, public_ip, "public_ip must not be empty")
}


