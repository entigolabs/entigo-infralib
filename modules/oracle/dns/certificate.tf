# One wildcard certificate for the whole zone, issued by Let's Encrypt and imported into
# the OCI Certificates service.
#
# Why a wildcard, and why here rather than in cert-manager like every other cloud:
# modules/k8s/oci-native-ingress-controller programs an OCI load balancer directly, and an
# OCI load balancer listener can carry exactly one key pair - there is no SNI. Every app
# served by NIC therefore has to present the same certificate, which forces both a wildcard
# and a single owner for it. Terraform is that owner because the certificate's OCID has to
# be known when the apps step renders its Helm values (the OCID goes into an Ingress
# annotation), which is before anything is running in the cluster.
#
# Only DNS-01 can issue a wildcard, so the challenge is solved against this very zone
# through lego's oraclecloud provider, embedded in the acme provider.

resource "tls_private_key" "acme_account" {
  count = var.create_certificate ? 1 : 0

  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "acme_registration" "this" {
  count = var.create_certificate ? 1 : 0

  account_key_pem = tls_private_key.acme_account[0].private_key_pem
  email_address   = var.acme_email
}

resource "acme_certificate" "wildcard" {
  count = var.create_certificate ? 1 : 0

  account_key_pem = acme_registration.this[0].account_key_pem
  # The apex is carried as a SAN so the zone root works too - a wildcard covers
  # foo.dev.example.com but never dev.example.com itself.
  common_name               = "*.${oci_dns_zone.pub.name}"
  subject_alternative_names = [oci_dns_zone.pub.name]

  # Renewal happens on whichever terraform run first sees the certificate inside this
  # window, so the deployment has to be applied at least this often. Let's Encrypt is
  # moving to 45-day lifetimes, which leaves a fortnight of slack at the default 30.
  min_days_remaining = var.certificate_min_days_remaining

  dns_challenge {
    provider = "oraclecloud"

    # Only the compartment is set here. Everything else lego needs (OCI_AUTH_TYPE, or the
    # OCI_TENANCY_OCID / OCI_USER_OCID / OCI_PUBKEY_FINGERPRINT / OCI_PRIVKEY_FILE set for
    # plain API-key auth, which also accept the TF_VAR_tenancy_ocid-style aliases) is read
    # from the environment, so the credentials never have to be written into a config file
    # that gets committed. Pass them through acme_dns_challenge_config only if the
    # environment cannot carry them.
    #
    # lego supports OCI_AUTH_TYPE=instance_principal, but NOT resource principal - so a run
    # from an OCI Container Instance (which is how the agent executes remotely) has to fall
    # back to API-key credentials. Verified against lego v4.35.2's oraclecloud provider,
    # the version vendored by acme 2.48.3.
    config = merge({
      OCI_COMPARTMENT_OCID = var.compartment_id
    }, var.acme_dns_challenge_config)
  }

  depends_on = [oci_dns_zone.pub]
}

# The certificate name cannot be changed after creation, and deleting an OCI certificate
# only *schedules* its deletion - the name stays taken for the duration. That would make a
# teardown/rebuild cycle collide with itself, so the name carries a random suffix. The
# keeper ties it to the domain rather than to the certificate content: re-issuing on
# renewal must reuse the same OCI certificate (a new version of it), because the OCID is
# baked into Ingress annotations that only get re-rendered on the next agent run.
resource "random_id" "certificate" {
  count = var.create_certificate ? 1 : 0

  byte_length = 4
  keepers = {
    domain = local.domain
  }
}

resource "oci_certificates_management_certificate" "wildcard" {
  count = var.create_certificate ? 1 : 0

  compartment_id = var.compartment_id
  name           = "${var.prefix}-wildcard-${random_id.certificate[0].hex}"
  description    = "Let's Encrypt wildcard for *.${oci_dns_zone.pub.name}, issued by modules/oracle/dns"

  certificate_config {
    config_type = "IMPORTED"
    # acme_certificate splits the PEM the way OCI wants it: certificate_pem is the leaf
    # only, issuer_pem is the chain above it.
    certificate_pem = acme_certificate.wildcard[0].certificate_pem
    cert_chain_pem  = acme_certificate.wildcard[0].issuer_pem
    private_key_pem = acme_certificate.wildcard[0].private_key_pem
  }

  lifecycle {
    # Renewal replaces the PEMs in place, which creates a new version of the same
    # certificate. Anything that would instead force a new certificate resource - and so a
    # new OCID - would silently strand every Ingress annotation still pointing at the old
    # one until the next apps run.
    create_before_destroy = true
  }
}
