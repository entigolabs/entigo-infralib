# Gateway configuration

## Default gateways and custom gateways

The module creates the following gateways by default, and additional gateways
can be defined by simply adding an entry under `gateways`. An entry is only
created when it sets `enabled: true`:

```yaml
gateways:
  external:
    enabled: true
    gatewayClassName: gke-l7-global-external-managed
    certificateMap: ''
    certManagerCerts: ''
    sslRedirect: true
  internal:
    enabled: true
    gatewayClassName: gke-l7-rilb
    certificateMap: ''
    certManagerCerts: ''
    sslRedirect: true
    allowGlobalAccess: false
  # custom user-defined gateway, just add an entry
  # partner:
  #   enabled: true
  #   gatewayClassName: gke-l7-global-external-managed
  #   certificateMap: "projects/my-project/locations/global/certificateMaps/partner"
  #   sslRedirect: true
```

Here the `partner` gateway would be created in addition to the defaults.

Each entry produces one `Gateway`, which is one GKE load balancer. The Gateway
object is named `<release name>-<key>`, so the entries above create
`google-gateway-external` and `google-gateway-internal`.

A custom entry must set its own `gatewayClassName`; the built-in entries come
with the class they have always used.

### No per-gateway subnets

Unlike the `aws-alb` module there is no `subnets` setting. The google `vpc`
module only supports the `default` subnet layout and has no `spoke` mode, so
there are no separate control/service subnets to place a gateway in. GKE picks
the subnet from the gateway class, which is why the class is the only placement
control here.

## Referencing a gateway from another module

The names of the two built-in gateways are published so other modules do not
have to hardcode them:

```yaml
global:
  internalGateway: google-gateway-internal
  externalGateway: google-gateway-external
```

The agent sets both from the module name, so they follow the module even when
it is deployed under a different name. A consuming module refers to them the
same way it does for `aws-alb`:

```yaml
global:
  google:
    gateway:
      name: '{{ .tinput.google-gateway.global.internalGateway }}'
      namespace: '{{ .tmodule.google-gateway }}'
```

Note that on google the value is the full Gateway object name, since gateways
here are prefixed with the release name.

## Certificates

The two GKE gateway classes take certificates in different ways, so there are
two settings:

- `certificateMap` renders the `networking.gke.io/certmap` annotation and is
  used by the global external classes backed by Certificate Manager.
- `certManagerCerts` renders `networking.gke.io/cert-manager-certs` on the
  https listener and is used by the regional internal classes.

Both are wired from the `dns` module by the agent for the built-in gateways:
the public zone for `external`, the internal zone for `internal`.

## HTTP to HTTPS redirect

Every gateway gets a catch-all `HTTPRoute` on its http listener that redirects
to https with a 301, named `<release name>-<key>-redirect`. Set
`sslRedirect: false` on an entry to leave plain HTTP in place.

## Global access

`allowGlobalAccess: true` creates a `GCPGatewayPolicy` that lets clients from
any region reach the load balancer. It only applies to regional internal
gateway classes such as `gke-l7-rilb`.

## SSL policies

`sslPolicy` sets the TLS versions and cipher suites the load balancer accepts
from clients, the google equivalent of the `aws-alb` `sslPolicy` setting. It is
per gateway, so each gateway can have its own. Leaving it unset keeps the GKE
default, which still permits TLS 1.0 and 1.1.

The policies themselves are created by the `dns` module, which makes three of
them, each existing both as a global and as a regional policy under the same
name:

| Output | Profile | Minimum TLS |
| --- | --- | --- |
| `ssl_policy_restricted` | `RESTRICTED` | 1.2 |
| `ssl_policy_modern` | `MODERN` | 1.2 |
| `ssl_policy_compatible` | `COMPATIBLE` | 1.0 |

`modern` is the default: the agent wires it into both built-in gateways, so
every gateway requires TLS 1.2 out of the box. `compatible` matches the google
default that still permits TLS 1.0, for a gateway that has to serve old clients.
`restricted` allows only AEAD ciphers, for stricter compliance requirements.

Because each name exists in both scopes, the same value works for a global
gateway class such as `gke-l7-global-external-managed` and for a regional one
such as `gke-l7-rilb` - the load balancer picks up the policy matching its own
scope. To use a non-default policy, point the gateway at another output:

```yaml
gateways:
  external:
    sslPolicy: '{{ .toptout.dns.ssl_policy_restricted }}'
  internal:
    sslPolicy: '{{ .toptout.dns.ssl_policy_compatible }}'
```

A policy managed elsewhere can be referenced by its plain name instead. Leaving
`sslPolicy` empty attaches no policy at all, which falls back to the GKE default
that permits TLS 1.0 and 1.1.

### One policy object per gateway

`allowGlobalAccess` and `sslPolicy` are both carried by the same
`GCPGatewayPolicy`, named `<release name>-<key>`, because GKE accepts only one
`GCPGatewayPolicy` per Gateway - a second one targeting the same Gateway is
rejected as a conflict and the oldest wins. Further gateway policy settings must
be added to that same object in `templates/gateways.yaml`, not as a new
document.
