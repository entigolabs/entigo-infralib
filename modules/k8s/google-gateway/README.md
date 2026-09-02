# Gateway configuration

## Default gateways and custom gateways

The module creates the following gateways by default, and additional gateways
can be defined by simply adding an entry under `gateways`:

```yaml
gateways:
  external:
    enabled: true
    gatewayClassName: gke-l7-global-external-managed
    certificateMap: ""
    certManagerCerts: ""
    sslRedirect: true
  internal:
    enabled: true
    gatewayClassName: gke-l7-rilb
    certificateMap: ""
    certManagerCerts: ""
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
      name: "{{ .tinput.google-gateway.global.internalGateway }}"
      namespace: "{{ .tmodule.google-gateway }}"
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

## Deprecated values

The previous fixed internal/external configuration is still honoured, so
existing configurations keep working unchanged:

| Deprecated value | Replacement |
| --- | --- |
| `global.createExternal` | `gateways.external.enabled` |
| `global.createInternal` | `gateways.internal.enabled` |
| `global.google.externalCertificateMap` | `gateways.external.certificateMap` |
| `global.google.internalCertificateMap` | `gateways.internal.certManagerCerts` |
| `global.google.externalGatewayClassName` | `gateways.external.gatewayClassName` |
| `global.google.internalGatewayClassName` | `gateways.internal.gatewayClassName` |
| `global.google.internalGatewayAllowGlobalAccess` | `gateways.internal.allowGlobalAccess` |

`createExternal` and `createInternal` default to `true` and are combined with
`enabled` using AND, so a built-in gateway is created only when both are true.

The rest default to empty, meaning unused. A configuration that still sets one
of them overrides the matching value in the `gateways` block, which is what
keeps older configurations working. To move to the `gateways` block, drop the
deprecated value — while it is set, the `gateways` value is ignored.
