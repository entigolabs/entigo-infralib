variable "master_authorized_networks" {
    type = list(object({
        cidr_block   = string
        display_name = string
    }))

    default = [
      {
        display_name = "Whitelist 1 - Entigo VPN"
        cidr_block   = "13.51.186.14/32"
      },
      {
        display_name = "Whitelist 2 - Entigo VPN"
        cidr_block   = "13.53.208.166/32"
      }
    ]
}