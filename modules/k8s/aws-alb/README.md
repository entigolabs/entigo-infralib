# Gateway configuration

## Default gateways and custom gateways

The module creates the following gateways by default if gateway creation is enabled and additional gateways can be defined by simply adding an entry under `gateways`:

```yaml
gateways:
  tags: {}
  internal:
    enabled: true
    scheme: internal
    sslRedirect: true
    subnets: ""
    certificates: ""
  service:
    enabled: false
    scheme: internal
    sslRedirect: true
    subnets: ""
    certificates: ""
  external:
    enabled: true
    scheme: internet-facing
    sslRedirect: true
    subnets: ""
    certificates: ""
  # custom user-defined gateway, just add an entry
  # partner:
  #   enabled: true
  #   scheme: internet-facing
  #   sslRedirect: true
  #   subnets: "subnet-aaa,subnet-bbb"
```

Here the `partner` gateway would be created in addition to the defaults.

### The service gateway

The `service` gateway is **disabled by default**. It can be enabled by simply
setting:

```yaml
gateways:
  service:
    enabled: true
```

The service gateway only makes sense when the `vpc` module `subnet_split_mode`
value is `spoke` — otherwise the internal and service gateway would both be
placed in the same private subnet.

In `spoke` mode the subnets are split by purpose:

- The **internal** gateway is attached to the **control subnets** and is meant
  for control plane applications such as ArgoCD.
- The **service** gateway is placed in the **service subnets**, meant for
  non-internet-accessible end user applications.

### Custom subnets

Subnets can be set explicitly as a comma-separated list (see `partner` above),
or fetched automatically from terraform outputs using the templating pattern:

```yaml
gateways:
  internal:
    subnets: '{{ .toutput.vpc.control_subnets }}'
  service:
    subnets: '{{ .toutput.vpc.service_subnets }}'
  external:
    subnets: '{{ .toutput.vpc.public_subnets }}'
```

### Selecting which gateway infralib modules use

Which gateway the infralib modules attach to for internal and external
services is configured centrally:

```yaml
global:
  internalGateway: internal
  externalGateway: external
```

If you want all internal services to use the external gateway, set
`global.internalGateway` to `external` (or any other gateway you have
defined, such as `partner`). Some infralib modules are external by default;
those are controlled by `global.externalGateway`.


# Migrating from Ingress to Gateway API (aws-alb module)

This guide describes how to migrate the `aws-alb` module and all workloads from
Kubernetes `Ingress` resources to Gateway API `Gateway` and `HTTPRoute` resources.

The migration is done in phases so that both routing stacks run in parallel,
allowing a gradual and controlled switchover of traffic.

## Overview of phases

| Phase | routingResources | createIngress | createRoute | State |
|-------|------------------|---------------|-------------|-------|
| 0. Initial | `ingress` (default) | `true` (default) | `false` (default) | Ingress only |
| 1. Enable gateways | `both` | `true` | `false` | Gateways created, Ingress still serving |
| 2. Enable routes for modules | `both` | `true` | `true` | Ingress and HTTPRoute in parallel |
| 3. Remove module ingresses | `both` | `false` | `true` | HTTPRoute serving, IngressClasses removed |
| 4. Finalize | `httproute` | `false` | `true` | Gateway API only |

## Phase 0: Initial state

A typical agent configuration for the `aws-alb` module is empty:

```yaml
      - name: aws-alb
        source: aws-alb
```

This is equivalent to the following default values:

```yaml
global:
  routingResources: "ingress"
  createIngress: true
  createRoute: false
  createType: ingress
```

In this state only the Ingress-based routing resources (ALB IngressClasses) exist.
During each phase the Ingralib Agent `run` command must be executed.

## Phase 1: Enable the Gateway objects

Set `routingResources` to `both` so the module creates the `GatewayClass` and
`Gateway` objects alongside the existing Ingress resources:

```yaml
      - name: aws-alb
        source: aws-alb
        inputs:
          global:
            routingResources: "both"
```

You might also want to specify what Gateways to create. For example if you do not use external `Gatway` and use a service `Gateway` then your configuration would be as follows:
```yaml
      - name: aws-alb
        source: aws-alb
        inputs:
          global:
            routingResources: "both"
          gateways:
            service:
              enabled: true
            external:
              enabled: false
```
This would provision internal and service `Gateway`. Without these settings internal and external `Gateway` would be provisioned.

Apply and wait until the Gateway load balancers are provisioned:

```bash
kubectl get gateway -A
```

Each enabled gateway should reach `PROGRAMMED: True` and have an address in
`status.addresses`.

## Phase 2: Enable HTTPRoutes for infralib modules

Once the load balancers exist, enable route creation for the other infralib
modules. Ingress and HTTPRoute objects will exist in parallel:

```yaml
      - name: aws-alb
        source: aws-alb
        inputs:
          global:
            routingResources: "both"
            createIngress: true
            createRoute: true
            createType: route
```

> **Note on `createType`:** `createType` is only used by the **harbor** module.
> It can be switched to `route` immediately.

Verify that the old `Ingress` objects are equivalent to the new `HTTPRoute` objects
```
kubectl get ingress -A
kubectl get httproute -A
```

> If for some modules the `Ingress` object is disabled then the `HTTProute` object needs to be disabled from a different input value as well.
> If the `Ingress` object used annotations for fine tuning then the related `HTTPRoute` inputs need to be changed as well or a `TargetGroupConfiguration` to be created for it.

## Phase 3: Remove the module Ingress objects

After the HTTPRoutes are created and DNS has been switched over to the new
Gateway load balancers, remove the Ingress objects of the infralib modules:

> **external-dns** will switch over the DNS automatically once Ingress objects are removed.
> If you want to migrate without interuption you need to make sure the DNS records point to the new loadbalancers before removing the Ingress objects.

```yaml
      - name: aws-alb
        source: aws-alb
        inputs:
          global:
            routingResources: "both" #ingress, both, httproute
            createIngress: false
            createRoute: true
            createType: route
```

This removes the Ingress objects of the infralib modules.

> If the DNS records were not changed then the `apps` step of the infralib agent will fail because it will lose connectivity with ArgoCD. Run the step again to fix the issue.

## Migrating your own applications

The steps above only cover infralib managed modules. **You are responsible for
converting the Ingress resources of your own applications to HTTPRoutes.**

The mapping rule is simple: the `ingressClassName` of your Ingress maps to the
Gateway of the same name. If your Ingress used ingressClass `external`, the
HTTPRoute attaches to Gateway `external`, ingressClass `internal` maps to
Gateway `internal`, and so forth. The Gateways live in the `aws-alb-<env>`
namespace.

### Example: converting an Ingress to an HTTPRoute

Existing Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  namespace: my-app-biz
spec:
  ingressClassName: external
  rules:
    - host: my-app.biz-net-route53.infralib.entigo.io
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 8080
```

Equivalent HTTPRoute:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-app
  namespace: my-app-biz
spec:
  parentRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: external          # was ingressClassName: external
      namespace: aws-alb-biz  # gateways live in the aws-alb-<env> namespace
      sectionName: https
  hostnames:
    - my-app.biz-net-route53.infralib.entigo.io   # was spec.rules[].host
  rules:
    - matches:
        - path:
            type: PathPrefix   # was pathType: Prefix
            value: /
      backendRefs:
        - group: ""
          kind: Service
          name: my-app         # was backend.service.name
          port: 8080           # was backend.service.port.number
```

Notes:

- One Ingress `host` becomes an entry in `spec.hostnames`. An Ingress with
  multiple hosts can be split into multiple HTTPRoutes or use multiple
  hostnames on one route.
- Each Ingress path entry becomes an entry in `spec.rules[].matches`.
- `pathType: Prefix` maps to `type: PathPrefix`, `pathType: Exact` maps to
  `type: Exact`.
- SSL redirect is handled on the Gateway (`sslRedirect: true`), so
  redirect-related Ingress annotations can be dropped.

### Example: multiple paths

```yaml
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /api
      backendRefs:
        - name: my-api
          port: 8080
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: my-frontend
          port: 80
```

## Phase 4: Finalize

Once all Ingress resources have been converted, verify that no Ingress objects
remain in the cluster:

```bash
kubectl get ingress -A
```

The command must return "No resources found". Then switch the module fully over to the
Gateway API:

```yaml
      - name: aws-alb
        source: aws-alb
        inputs:
          global:
            routingResources: "httproute" #ingress, both, httproute
            createIngress: false
            createRoute: true
            createType: route
```
