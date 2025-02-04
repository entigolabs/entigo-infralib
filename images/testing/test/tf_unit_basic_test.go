package test

import (
	"testing"
	_ "github.com/gruntwork-io/terratest/modules/aws"
	_ "github.com/gruntwork-io/terratest/modules/gcp"
	_ "github.com/gruntwork-io/terratest/modules/k8s"
	_ "github.com/gruntwork-io/terratest/modules/terraform"
	_ "github.com/gruntwork-io/terratest/modules/test-structure"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
	_ "github.com/gruntwork-io/terratest/modules/random"
)

func TestTerraformBasicOne(t *testing.T) {
	t.Log("This test only exists to cache correct dependencies")
}
