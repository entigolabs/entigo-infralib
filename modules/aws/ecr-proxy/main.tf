locals {
  hub = {
    username = var.hub_username
    accessToken = var.hub_token
  }
  ghcr = {
    username = var.ghcr_username
    accessToken = var.ghcr_token
  }
  gcr = {
    username = var.gcr_username
    accessToken = var.gcr_token
  }
}

resource "aws_secretsmanager_secret" "ecr_pullthroughcache_hub" {
  count = var.hub_username != "" && var.hub_token != "" ? 1 : 0
  name = "ecr-pullthroughcache/${substr(var.prefix, 0, 24)}-hub"
  recovery_window_in_days = 7
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    created-by = "entigo-infralib"
  }
}

resource "aws_secretsmanager_secret_version" "ecr_pullthroughcache_hub" {
  count = var.hub_username != "" && var.hub_token != "" ? 1 : 0
  secret_id     = aws_secretsmanager_secret.ecr_pullthroughcache_hub[0].id
  secret_string = jsonencode(local.hub)
}

resource "aws_secretsmanager_secret" "ecr_pullthroughcache_ghcr" {
  count = var.ghcr_username != "" && var.ghcr_token != "" ? 1 : 0
  name = "ecr-pullthroughcache/${substr(var.prefix, 0, 24)}-ghcr"
  recovery_window_in_days = 7
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    created-by = "entigo-infralib"
  }
}

resource "aws_secretsmanager_secret_version" "ecr_pullthroughcache_ghcr" {
  count = var.ghcr_username != "" && var.ghcr_token != "" ? 1 : 0
  secret_id     = aws_secretsmanager_secret.ecr_pullthroughcache_ghcr[0].id
  secret_string = jsonencode(local.ghcr)
}

resource "aws_secretsmanager_secret" "ecr_pullthroughcache_gcr" {
  count = var.gcr_username != "" && var.gcr_token != "" ? 1 : 0
  name = "ecr-pullthroughcache/${substr(var.prefix, 0, 24)}-gcr"
  recovery_window_in_days = 7
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    created-by = "entigo-infralib"
  }
}

resource "aws_secretsmanager_secret_version" "ecr_pullthroughcache_gcr" {
  count = var.gcr_username != "" && var.gcr_token != "" ? 1 : 0
  secret_id     = aws_secretsmanager_secret.ecr_pullthroughcache_gcr[0].id
  secret_string = jsonencode(local.gcr)
}

resource "aws_ecr_pull_through_cache_rule" "hub" {
  count = var.hub_username != "" && var.hub_token != "" ? 1 : 0
  ecr_repository_prefix = "${substr(var.prefix, 0, 24)}-hub"
  upstream_registry_url = "registry-1.docker.io"
  credential_arn        = aws_secretsmanager_secret.ecr_pullthroughcache_hub[0].arn
  depends_on = [
    aws_secretsmanager_secret_version.ecr_pullthroughcache_hub
  ]
}

resource "aws_ecr_pull_through_cache_rule" "ghcr" {
  count = var.ghcr_username != "" && var.ghcr_token != "" ? 1 : 0
  ecr_repository_prefix = "${substr(var.prefix, 0, 24)}-ghcr"
  upstream_registry_url = "ghcr.io"
  credential_arn        = aws_secretsmanager_secret.ecr_pullthroughcache_ghcr[0].arn
  depends_on = [
    aws_secretsmanager_secret_version.ecr_pullthroughcache_ghcr
  ]
}

resource "aws_ecr_pull_through_cache_rule" "gcr" {
  count = var.gcr_username != "" && var.gcr_token != "" ? 1 : 0
  ecr_repository_prefix = "${substr(var.prefix, 0, 24)}-gcr"
  upstream_registry_url = "gcr.io"
  credential_arn        = aws_secretsmanager_secret.ecr_pullthroughcache_gcr[0].arn
  depends_on = [
    aws_secretsmanager_secret_version.ecr_pullthroughcache_gcr
  ]
}

resource "aws_ecr_pull_through_cache_rule" "k8s" {
  ecr_repository_prefix = "${substr(var.prefix, 0, 24)}-k8s"
  upstream_registry_url = "registry.k8s.io"
}

resource "aws_ecr_pull_through_cache_rule" "ecr" {
  ecr_repository_prefix = "${substr(var.prefix, 0, 24)}-ecr"
  upstream_registry_url = "public.ecr.aws"
}

resource "aws_ecr_pull_through_cache_rule" "quay" {
  ecr_repository_prefix = "${substr(var.prefix, 0, 24)}-quay"
  upstream_registry_url = "quay.io"
}

resource "aws_ecr_repository_creation_template" "ecr-proxy" {
  for_each = toset(["hub", "ghcr", "gcr", "k8s", "ecr", "quay"])
  prefix               = "${substr(var.prefix, 0, 24)}-${each.value}"
  description          = "${var.prefix}-${each.value}"
  image_tag_mutability = "MUTABLE"

  applied_for = [
    "PULL_THROUGH_CACHE",
  ]

  encryption_configuration {
    encryption_type = "AES256"
  }

  lifecycle_policy = <<EOT
{
  "rules": [
        {
            "rulePriority": 1,
            "description": "Expire untagged images older than 7 days",
            "selection": {
                "tagStatus": "untagged",
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 7
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 2,
            "description": "Expire tagged images older than 90 days",
            "selection": {
                "tagStatus": "tagged",
                "tagPatternList": ["*"],
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 90
            },
            "action": {
                "type": "expire"
            }
        }
  ]
}
EOT

}


resource "aws_iam_policy" "ecr-proxy" {
  name        = substr(var.prefix, 0, 24)
  path        = "/"
  description = "ECR ${substr(var.prefix, 0, 24)} usage"

  tags = {
    Terraform   = "true"
    Environment = var.prefix
    created-by = "entigo-infralib"
  }

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:DescribeRepositories",
                "ecr:ListImages",
                "ecr:DescribeImages",
                "ecr:BatchGetImage",
                "ecr:ListTagsForResource",
                "ecr:DescribeImageScanFindings",
                "ecr:ReplicateImage",
                "ecr:CreateRepository",
                "ecr:BatchImportUpstreamImage",
                "ecr:TagResource"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/${substr(var.prefix, 0, 24)}-*"
      },
    ]
  })
}
