package test

import (
	"testing"
	"fmt"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/stretchr/testify/assert"
)



func TestTerraformHelloWorld(t *testing.T) {
	t.Run("Biz", testTerraformHelloWorldBiz)
	t.Run("Pri", testTerraformHelloWorldPri)
}

func testTerraformHelloWorldBiz(t *testing.T) {
        t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz", "net")
	hello_world := aws.GetStringValue(t, outputs, "hello-world__hello_world")
	assert.Equal(t, hello_world, fmt.Sprintf("Hello, biz-net-hello-world!"))
  
}

func testTerraformHelloWorldPri(t *testing.T) {
        t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri", "net")
	hello_world := aws.GetStringValue(t, outputs, "hello-world__hello_world")
	assert.Equal(t, hello_world, fmt.Sprintf("Hello, pri-net-hello-world!"))
}
