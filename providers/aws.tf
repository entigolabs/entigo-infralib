provider "aws" {
  ignore_tags {
      key_prefixes = ["kubernetes.io/cluster/"]
      keys = []
  }
  default_tags {
    tags = {}
  }
}

# tflint-ignore: terraform_unused_declarations
provider "aws" {
  region = "us-east-1"
  alias  = "us-east-1"
  ignore_tags {
      key_prefixes = []
      keys = []
  }
  default_tags {
    tags = {}
  }
}

