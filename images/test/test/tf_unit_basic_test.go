package test

// This test only exists to warm the Go module and build caches of the image.
// Import every package the module tests under modules/*/*/test import, plus
// the shared library packages from common/, so nothing is downloaded or
// compiled from scratch when a module test runs. Keep it in sync with:
//   grep -rhE '^\s*([A-Za-z0-9_]+\s+)?"[a-z0-9.-]+\.[a-z]+/' modules --include=*_test.go | sort -u

import (
	"testing"

	_ "github.com/entigolabs/entigo-infralib-common/aws"
	_ "github.com/entigolabs/entigo-infralib-common/google"
	_ "github.com/entigolabs/entigo-infralib-common/k8s"
	_ "github.com/entigolabs/entigo-infralib-common/oracle"
	_ "github.com/entigolabs/entigo-infralib-common/tf"
	_ "github.com/gruntwork-io/terratest/modules/aws"
	_ "github.com/gruntwork-io/terratest/modules/gcp"
	_ "github.com/gruntwork-io/terratest/modules/k8s"
	_ "github.com/gruntwork-io/terratest/modules/logger"
	_ "github.com/gruntwork-io/terratest/modules/random"
	_ "github.com/gruntwork-io/terratest/modules/retry"
	_ "github.com/gruntwork-io/terratest/modules/terraform"
	_ "github.com/gruntwork-io/terratest/modules/test-structure"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
	_ "k8s.io/api/core/v1"
	_ "k8s.io/api/storage/v1"
	_ "k8s.io/apimachinery/pkg/api/resource"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	_ "k8s.io/apimachinery/pkg/runtime/schema"
)

func TestTerraformBasicOne(t *testing.T) {
	t.Log("This test only exists to cache correct dependencies")
}
