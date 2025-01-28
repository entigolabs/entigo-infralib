package test

import (
	"testing"
	"github.com/entigolabs/entigo-infralib-common/tf"
	"github.com/entigolabs/entigo-infralib-common/google"
	"github.com/stretchr/testify/assert"
)

func TestTerraformServices(t *testing.T) {
       outputs := google.GetTFOutputs(t, "biz", "net")
      services := tf.GetStringValue(t, outputs, "services__services")
      assert.NotEmpty(t, services, "services was not returned")
}
