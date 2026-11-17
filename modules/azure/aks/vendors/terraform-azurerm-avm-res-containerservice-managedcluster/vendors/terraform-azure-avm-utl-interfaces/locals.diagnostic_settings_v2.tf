# These locals create the data for the azapi_resources from the var.diagnostic_settings_v2 variable
locals {
  diagnostic_settings_azapi_v2 = {
    for k, v in var.diagnostic_settings_v2 : k => {
      type = local.diagnostic_settings_type_v2
      name = v.name
      body = {
        properties = {
          eventHubAuthorizationRuleId = lookup(v, "event_hub_authorization_rule_resource_id", null)
          eventHubName                = lookup(v, "event_hub_name", null)
          logAnalyticsDestinationType = lookup(v, "log_analytics_destination_type", null)
          logs = length(v.logs) > 0 ? [
            for log in v.logs :
            {
              category      = log.category
              categoryGroup = log.category_group
              enabled       = log.enabled
              retentionPolicy = {
                days    = log.retention_policy.days
                enabled = log.retention_policy.enabled
              }
            }
          ] : null
          marketplacePartnerId = lookup(v, "marketplace_partner_resource_id", null)
          metrics = length(v.metrics) > 0 ? [
            for metric in v.metrics :
            {
              category = metric.category
              enabled  = metric.enabled
              retentionPolicy = {
                days    = metric.retention_policy.days
                enabled = metric.retention_policy.enabled
              }
            }
          ] : null
          storageAccountId = lookup(v, "storage_account_resource_id", null)
          workspaceId      = lookup(v, "workspace_resource_id", null)
        }
      }
    }
  }
  diagnostic_settings_type_v2 = "Microsoft.Insights/diagnosticSettings@2021-05-01-preview"
}
