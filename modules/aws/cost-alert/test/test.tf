provider "aws" {
  region = "us-east-1"
  alias  = "us-east-1"
}

module "cost_alert_test" {
  source = "../"

  providers = {
    aws = aws.us-east-1
  }

  prefix                    = var.prefix
  monthly_billing_threshold = var.monthly_billing_threshold
  alert_emails              = var.alert_emails
}

output "sns_topic_arns" {
  value = module.cost_alert_test.sns_topic_arns
}
