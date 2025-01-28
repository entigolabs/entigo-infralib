package test

import (
	"testing"
	"fmt"
	"os"
	"strings"
	"github.com/entigolabs/entigo-infralib-common/aws"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/stretchr/testify/assert"
)



func TestTerraformHelloWorld(t *testing.T) {
	t.Run("Biz", testTerraformHelloWorldBiz)
	t.Run("Pri", testTerraformHelloWorldPri)
}

func testTerraformHelloWorldBiz(t *testing.T) {
        t.Parallel()
	outputs := aws.GetTFOutputs(t, "biz")
	hello_world := tf.GetStringValue(t, outputs, "hello-world__hello_world")
	assert.Equal(t, hello_world, fmt.Sprintf("Hello, biz-%s-hello-world!", strings.ToLower(os.Getenv("STEP_NAME"))))
  
}

func testTerraformHelloWorldPri(t *testing.T) {
        t.Parallel()
	outputs := aws.GetTFOutputs(t, "pri")
	hello_world := tf.GetStringValue(t, outputs, "hello-world__hello_world")
	assert.Equal(t, hello_world, fmt.Sprintf("Hello, pri-%s-hello-world!", strings.ToLower(os.Getenv("STEP_NAME"))))
}

