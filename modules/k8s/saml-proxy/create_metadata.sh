#!/usr/bin/env bash
set -e

PROG="$(basename "$0")"

printUsage() {
    echo "Usage: $PROG ENTITY-ID ENDPOINT-URL"
    echo ""
    echo "Example:"
    echo "  $PROG urn:someservice https://sp.example.org/mellon"
    echo ""
}

if [ "$#" -lt 2 ]; then
    printUsage
    exit 1
fi

ENTITYID="$1"
if [ -z "$ENTITYID" ]; then
    echo "$PROG: An entity ID is required." >&2
    exit 1
fi

BASEURL="$2"
if [ -z "$BASEURL" ]; then
    echo "$PROG: The URL to the MellonEndpointPath is required." >&2
    exit 1
fi

AWS_SM_KEY="$3"
if [ -z "$AWS_SM_KEY" ]; then
    echo "$PROG: The AWS SecretManager secret name is required." >&2
    exit 1
fi

AWS_REGION="$4"
if [ -z "$AWS_REGION" ]; then
    echo "$PROG: The AWS region is required." >&2
    exit 1
fi

if ! echo "$BASEURL" | grep -q '^https\?://'; then
    echo "$PROG: The URL must start with \"http://\" or \"https://\"." >&2
    exit 1
fi

HOST="$(echo "$BASEURL" | sed 's#^[a-z]*://\([^/]*\).*#\1#')"
BASEURL="$(echo "$BASEURL" | sed 's#/$##')"

OUTFILE="saml_sp"
echo "Output files:"
echo "Private key:               $OUTFILE.key"
echo "Certificate:               $OUTFILE.cert"
echo "Metadata:                  $OUTFILE.xml"
echo "Host:                      $HOST"
echo
echo "Endpoints:"
echo "SingleLogoutService:       $BASEURL/logout"
echo "AssertionConsumerService:  $BASEURL/postResponse"
echo

# No files should not be readable by the rest of the world.
umask 0077

RANDFILE="$(mktemp -t mellon_rndfile.XXXXXXXXXX)"
TEMPLATEFILE="$(mktemp -t mellon_create_sp.XXXXXXXXXX)"

openssl rand -out $RANDFILE -hex 256

cat >"$TEMPLATEFILE" <<EOF
RANDFILE           = $RANDFILE
[req]
default_bits       = 2048
default_keyfile    = privkey.pem
distinguished_name = req_distinguished_name
prompt             = no
policy             = policy_anything
[req_distinguished_name]
commonName         = $HOST
EOF

openssl req -utf8 -batch -config "$TEMPLATEFILE" -new -x509 -days 3652 -nodes -out "$OUTFILE.cert" -keyout "$OUTFILE.key" 2>/dev/null

rm -f "$RANDFILE"
rm -f "$TEMPLATEFILE"

CERT="$(grep -v '^-----' "$OUTFILE.cert")"

cat >"$OUTFILE.xml" <<EOF
<EntityDescriptor entityID="$ENTITYID" xmlns="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
  <SPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <KeyDescriptor use="signing">
      <ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
        <ds:X509Data>
          <ds:X509Certificate>$CERT</ds:X509Certificate>
        </ds:X509Data>
      </ds:KeyInfo>
    </KeyDescriptor>
    <SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="$BASEURL/logout"/>
    <AssertionConsumerService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="$BASEURL/postResponse" index="0"/>
    <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</NameIDFormat>
  </SPSSODescriptor>
</EntityDescriptor>
EOF

umask 0777
chmod go+r "$OUTFILE.xml"
chmod go+r "$OUTFILE.cert"


touch secrets.json
chmod 666 secrets.json
jq -n \
  --arg saml_idp_xml "$(cat saml_idp.xml)" \
  --arg saml_sp_cert "$(cat saml_sp.cert)" \
  --arg saml_sp_key "$(cat saml_sp.key)" \
  --arg saml_sp_xml "$(cat saml_sp.xml)" \
  '{ "saml_idp.xml": $saml_idp_xml, "saml_sp.cert": $saml_sp_cert, "saml_sp.key": $saml_sp_key, "saml_sp.xml": $saml_sp_xml }' > secrets.json

aws secretsmanager create-secret --name "$AWS_SM_KEY" --region "$AWS_REGION" --secret-string file://secrets.json || aws secretsmanager update-secret --secret-id "$AWS_SM_KEY" --region "$AWS_REGION" --secret-string file://secrets.json

rm -f "$OUTFILE.xml"
rm -f "$OUTFILE.cert"
rm -f "$OUTFILE.key"
rm -f secrets.json
