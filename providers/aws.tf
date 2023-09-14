provider "aws" {
  ignore_tags {
      key_prefixes = ["kubernetes.io/cluster/"]
  }
  default_tags {
  }
}

provider "aws" {
  region = "us-east-1"
  alias  = "us-east-1"
  ignore_tags {
  }
  default_tags {
  }
}
