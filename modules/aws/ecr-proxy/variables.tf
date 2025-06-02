variable "prefix" {
  type = string
}

variable "hub_username" {
  type = string
  sensitive   = true
  default = ""
}

variable "hub_token" {
  type = string
  sensitive   = true
  default = ""
}

variable "ghcr_username" {
  type = string
  sensitive   = true
  default = ""
}

variable "ghcr_token" {
  type = string
  sensitive   = true
  default = ""
}

variable "gcr_username" {
  type = string
  sensitive   = true
  default = ""
}

variable "gcr_token" {
  type = string
  sensitive   = true
  default = ""
}

variable "upstream_registry_url" {
  type = string
  default = ""
}

variable "upstream_registry_lifecycle_policy" {
  type = string
  default = <<EOT
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
