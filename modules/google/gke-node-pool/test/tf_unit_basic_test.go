package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/google"
)

func TestTerraformGke(t *testing.T) {
	t.Run("Biz", testTerraformGkeBiz)
}

func testTerraformGkeBiz(t *testing.T) {
	t.Parallel()
	testTerraformGke(t, "biz")
}

func testTerraformGke(t *testing.T, envName string) {
	outputs := google.GetTFOutputs(t, envName)

	_ = outputs
}
