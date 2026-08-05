provider "acme" {
  # Used by modules/oracle/dns to issue the zone-wide wildcard certificate that
  # modules/k8s/oci-native-ingress-controller serves every host from.
  #
  # server_url is a required provider argument (it falls back to ACME_SERVER_URL, but only
  # when omitted entirely), so it is pinned here rather than left to an environment variable
  # the agent has no reason to pass through. To issue against Let's Encrypt's staging CA
  # while testing - untrusted certificates, but far looser rate limits - change this to
  # https://acme-staging-v02.api.letsencrypt.org/directory and re-run. Note that the
  # per-registered-domain issuance limit is counted against entigo.dev as a whole, shared
  # with everyone else in the organisation, so repeated teardown/rebuild cycles against
  # production are worth avoiding.
  server_url = "https://acme-v02.api.letsencrypt.org/directory"
}
