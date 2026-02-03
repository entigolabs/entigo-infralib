package test

import (
	"testing"

	"github.com/entigolabs/entigo-infralib-common/google"
)

func TestTerraformKms(t *testing.T) {
	t.Run("Biz", testTerraformKmsBiz)
}

func testTerraformKmsBiz(t *testing.T) {
	t.Parallel()
	outputs := google.GetTFOutputs(t, "biz")
	_ = outputs
}
