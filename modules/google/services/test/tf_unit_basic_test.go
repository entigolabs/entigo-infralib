package test

import (
	"testing"

	commonGoogle "github.com/entigolabs/entigo-infralib-common/google"
	"github.com/entigolabs/entigo-infralib-common/tf"
)

const bucketName = "infralib-modules-gce-services-tf"

var Region string

func TestTerraformServices(t *testing.T) {
	Region = commonGoogle.SetupBucket(t, bucketName)
	options := tf.InitGCloudTerraform(t, bucketName, Region, "tf_unit_basic_test_biz.tfvars", map[string]interface{}{})
	outputs, destroyFunc := tf.ApplyTerraform(t, "biz", options)
	defer destroyFunc()
	_ = outputs
}
