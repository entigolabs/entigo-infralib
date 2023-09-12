provider "aws" {
  ignore_tags {
      key_prefixes = ["kubernetes.io/cluster/"]
  }
}

provider "aws" {
  region = "us-east-1"
  alias  = "us-east-1"
}
