locals {
  hname  = "${var.prefix}-${terraform.workspace}"
  alarms = var.create_sns_topic ? concat([aws_sns_topic.this[0].arn], var.sns_topic_arns) : var.sns_topic_arns
}

# Alarm
#https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/monitor_estimated_charges_with_cloudwatch.html
resource "aws_cloudwatch_metric_alarm" "this" {
  alarm_name          = var.aws_account_id == null ? "global-billing-alarm-${lower(var.currency)}" : "account-billing-alarm-${lower(var.currency)}-${var.aws_account_id}"
  alarm_description   = var.aws_account_id == null ? "Billing consolidated alarm >= ${var.currency} ${var.monthly_billing_threshold}" : "Billing alarm account ${var.aws_account_id} >= ${var.currency} ${var.monthly_billing_threshold}"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "1"
  datapoints_to_alarm = "1"
  metric_name         = "EstimatedCharges"
  namespace           = "AWS/Billing"
  period              = "21600"
  statistic           = "Maximum"
  threshold           = var.monthly_billing_threshold
  alarm_actions       = local.alarms
  dimensions = {
    Currency      = var.currency
    LinkedAccount = var.aws_account_id
  }
  tags = {
    Terraform   = "true"
    Environment = local.hname
  }
}

# SNS Topic
resource "aws_sns_topic" "this" {
  count = var.create_sns_topic ? 1 : 0

  name = var.aws_account_id == null ? "global-billing-alarm-notification-${lower(var.currency)}" : "billing-alarm-notification-${lower(var.currency)}-${var.aws_account_id}"
  tags = {
    Terraform   = "true"
    Environment = local.hname
  }
}


resource "aws_sns_topic_subscription" "this" {
  for_each = var.create_sns_topic ? toset(var.alert_emails) : []

  topic_arn = aws_sns_topic.this[0].arn
  protocol  = "email"
  endpoint  = each.key
}
