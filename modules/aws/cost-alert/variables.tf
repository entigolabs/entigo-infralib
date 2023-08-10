variable "prefix" {
  type = string
}

variable "monthly_billing_threshold" {
  description = "The threshold for which estimated monthly charges will trigger the metric alarm."
  type        = string
}

variable "alert_emails" {
  description = "E-mails to send alerts to if `create_sns_topic` is `true`"
  type        = list(string)
  default     = []
}

variable "currency" {
  description = "Short notation for currency type (e.g. USD, CAD, EUR)"
  type        = string
  default     = "USD"
}

variable "aws_account_id" {
  description = "AWS account id"
  type        = string
  default     = null
}

variable "create_sns_topic" {
  description = "Creates a SNS Topic if `true`."
  type        = bool
  default     = true
}


variable "sns_topic_arns" {
  description = "List of SNS topic ARNs to be used. If `create_sns_topic` is `true`, it merges the created SNS Topic by this module with this list of ARNs"
  type        = list(string)
  default     = []
}

