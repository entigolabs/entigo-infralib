package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/google"
)

func TestTerraformGkeNodePool(t *testing.T) {
	t.Run("Pri", testTerraformGkeNodePoolPri)
	t.Run("Biz", testTerraformGkeNodePoolBiz)
}

func testTerraformGkeNodePoolPri(t *testing.T) {
	t.Parallel()
	testTerraformGkeNodePool(t, "pri")
}

func testTerraformGkeNodePoolBiz(t *testing.T) {
	t.Parallel()
	testTerraformGkeNodePool(t, "biz")
}

func testTerraformGkeNodePool(t *testing.T, envName string) {
	outputs := google.GetTFOutputs(t, envName)

	_ = outputs
}
