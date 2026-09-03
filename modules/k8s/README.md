# Kubernetes modules

Every folder under `modules/k8s/` is one module: a wrapper Helm chart around an
upstream chart, together with the cloud specific resources that chart needs and
the agent inputs that wire it to the rest of the platform.

Modules are deployed by the
[Infralib Agent](https://github.com/entigolabs/entigo-infralib-agent) in a step
of type `argocd-apps`. The agent renders one ArgoCD `Application` per module and
lets ArgoCD do the installing.

## Using a module

```yaml
steps:
  - name: apps
    type: argocd-apps
    modules:
      - name: hello-world
        source: hello-world
```

`source` is the folder name under `modules/k8s/`. `name` is what the module is
called in this platform, and it becomes both the ArgoCD `Application` name and
the namespace the module is installed into. The two are usually the same, but
they do not have to be — the same source can be deployed twice under different
names:

```yaml
      - name: grafana-team-a
        source: grafana
      - name: grafana-team-b
        source: grafana
```

Anything you put under `inputs:` is merged on top of the module's own values and
overrides them.

## Anatomy of a module

| File | Purpose |
|---|---|
| `Chart.yaml` | Wrapper chart metadata and the upstream chart dependency |
| `Chart.lock` | Resolved dependency digest, written by `helm dependency update` |
| `charts/*.tgz` | The vendored upstream chart, committed to the repository |
| `values.yaml` | Cloud neutral defaults for the wrapper and the upstream chart |
| `values-aws.yaml` / `values-google.yaml` | Defaults that only apply on that cloud |
| `agent_input.yaml` | Agent templated values, the environment specific wiring |
| `agent_input_aws.yaml` / `agent_input_google.yaml` | Wiring that exists on one cloud only |
| `templates/` | Resources this repository owns, not the upstream chart |
| `templates/aws/`, `templates/google/` | Cloud specific resources |
| `argo-apps.yaml` | Overrides merged into the generated ArgoCD `Application` |
| `agent.yaml` | Optional module metadata surfaced by the agent (UI URL, module type) |
| `test/` | Test environment inputs and Go tests, see [Tests](#tests) |
| `test.sh` | Entry point that runs the static and unit tests |
| `README.md` | Module specific documentation, when the module needs it |

A module that only ships its own manifests, for example `hello-world`,
`rbac-bindings`, `aws-storageclass`, `google-gateway` or the `crossplane-*`
provider modules, has no `charts/` folder and no dependencies in `Chart.yaml`.
Everything else in the table still applies.

## How values reach the cluster

This is the single most important thing to understand before editing a module.
The agent generates roughly this `Application`:

```yaml
spec:
  destination:
    namespace: '<module name>'
  sources:
    - repoURL: '<infralib repo>'
      targetRevision: '<module version>'
      path: "modules/k8s/<module source>"
      helm:
        ignoreMissingValueFiles: true
        valueFiles:
          - 'values.yaml'
          - 'values-<cloud>.yaml'
        values: |
          <the rendered agent inputs>
```

So two different mechanisms deliver values, and they do not have equal weight.
The value files are read by ArgoCD straight from this repository at the module's
version. The agent inputs are merged and templated by the agent and written into
the `Application` as inline Helm values, which in Helm always beat value files.

From weakest to strongest:

| # | Source | Who supplies it |
|---|---|---|
| 1 | `values.yaml` | this repository |
| 2 | `values-<cloud>.yaml` | this repository |
| 3 | `agent_input.yaml` | this repository, templated by the agent |
| 4 | `agent_input_<cloud>.yaml` | this repository, templated by the agent |
| 5 | `inputs:` in the agent config | the platform being deployed |

The dividing line between 1–2 and 3–5 is the principle to follow: **value files
hold what is true for every installation of this module, agent inputs hold what
can only be known once the module is placed in a real environment.** A chart
default, a resource limit, a security context or a hardcoded image tag belongs
in `values.yaml`. An account ID, a cluster OIDC issuer, a bucket name, a domain
or the name of another module belongs in `agent_input.yaml`.

Keeping that line clean is also what makes the static tests meaningful, because
they render the chart with the value files and nothing else.

### values.yaml

Cloud neutral defaults. Top level keys are `global:` plus one key per subchart,
named exactly as the dependency is named in `Chart.yaml`:

```yaml
global:
  cloudProvider: ""
  providerConfigRefName: ""

grafana:            # matches dependencies[].name in Chart.yaml
  enabled: true
```

Declare every key the templates read, even when the real value always arrives
from the agent. An undeclared key makes `helm template` fail in the static tests
and gives users of the module nothing to discover.

### values-aws.yaml and values-google.yaml

Only what genuinely differs per cloud, and always including `cloudProvider`:

```yaml
global:
  cloudProvider: "aws"                   # or "google"
  providerConfigRefName: "crossplane-aws"

grafana:
  persistence:
    storageClassName: "gp3"              # "standard" on google
```

`ignoreMissingValueFiles: true` means a module that only supports one cloud can
simply omit the other file.

`global.cloudProvider` is set here and nowhere else. It is what every cloud
specific template gates on, so it must never be templated by the agent.

## Agent inputs

`agent_input.yaml` is a Helm values file with agent replacement tags in it. The
agent resolves the tags while generating the `Application`, so the file is where
a module reaches out to the rest of the platform: Terraform outputs of the
infrastructure steps, and the inputs of the other modules deployed alongside it.

### Replacement tags

The tags used most in these modules:

| Tag | Resolves to |
|---|---|
| `.module.name` | this module's name |
| `.config.prefix` | the platform prefix |
| `.toutput.<source>.<key>` | a Terraform output of the module with that source |
| `.toptout.<source>.<key>` | the same, empty string when absent |
| `.tinput.<source>.<path>` | a value from another module in this step |
| `.toptin.<source>.<path>` | the same, empty string when absent |
| `.tmodule.<source>` | the name given to another module in this step |
| `.toptmodule.<source>` | the same, empty string when absent |

`<source>` is the module's source, not its name — `aws-alb`, `route53`,
`google-gateway`. Use `.tmodule` / `.toptmodule` when you need the name that
module was actually given, which is what its namespace is called.

`.tinput` and `.toptin` read the target module's merged values, so they see its
`values.yaml` and `values-<cloud>.yaml` too, not only its agent inputs. That is
why `{{ .toptin.aws-alb.global.internalGateway }}` works without anyone setting
`internalGateway` in the platform config — the default in `aws-alb/values.yaml`
answers it.

The full tag reference lives in the
[agent README](https://github.com/entigolabs/entigo-infralib-agent#overriding-config-values).

### Optional chains and `required`

An optional tag can be chained with `|`, and the first value that resolves to a
non-empty string wins. A quoted string at the end is a literal default:

```yaml
image:
  registry: '{{ .toptout.ecr-proxy.hub_registry | .toptout.gar-proxy.hub_registry | "docker.io" }}'
```

Ending the chain with `required` instead makes the agent fail — naming the whole
tag — when nothing in the chain resolved:

```yaml
grafana:
  route:
    main:
      parentRefs:
        - group: gateway.networking.k8s.io
          kind: Gateway
          name: "{{ .toptin.aws-alb.global.internalGateway | .toptin.google-gateway.global.internalGateway | required }}"
          namespace: "{{ .toptmodule.aws-alb | .toptmodule.google-gateway | required }}"
          sectionName: https
  grafana.ini:
    server:
      root_url: https://{{ .module.name }}.{{ .toptout.route53.int_domain | .toptout.dns.int_domain | required }}
```

This is what lets one file serve every cloud. On AWS the `aws-alb` and `route53`
links resolve and the google ones are empty; on Google it is the other way
round; on a platform that has neither, the deployment stops with a clear error
instead of quietly rendering an empty gateway name.

`required` must be last in the chain, must be preceded by at least one value,
and cannot be combined with a literal default — a default always resolves, so
the two together would be meaningless.

### One file, not two

Write the wiring in `agent_input.yaml` and use chains for the parts that differ.
Reach for `agent_input_aws.yaml` or `agent_input_google.yaml` only when
something exists on one cloud and has no counterpart on the other — an IAM role
built from `{{ .toutput.eks.oidc_provider }}`, a GCS bucket location, a
Workload Identity service account.

The per cloud files are merged on top of `agent_input.yaml`, so a key set in
both is decided by the cloud specific one.

Two things do not chain and still need splitting or restructuring:

- **Lists.** Maps are deep merged, lists are replaced whole. A list whose only
  cloud specific part is one field inside one element cannot be chained; either
  lift the differing field out of the list, or accept the duplication and keep
  the two copies in step.
- **Whole blocks that only exist on one cloud.** Chain the values, not the
  block.

## Cloud specific resources

Resources this repository owns live in `templates/`. A module that supports more
than one cloud puts the cloud specific ones in a folder named after the cloud:

```
templates/
  configMap.yaml          # rendered everywhere
  aws/
    Role.yaml
    serviceAccount.yaml
    targetGroupConfiguration.yaml
  google/
    serviceAccount.yaml
    healthCheckPolicy.yaml
```

**The folder is organisation, not logic.** Helm renders every file under
`templates/` regardless of which folder it is in, so each cloud specific file
must also gate itself:

```yaml
{{- if eq .Values.global.cloudProvider "aws" }}
...
{{- end }}
```

Without the gate the resource is rendered on every cloud, and in the static
tests — where `cloudProvider` is empty — it is rendered with empty values.

Modules that only ever run on one cloud (`karpenter`, `cluster-autoscaler`,
`aws-alb`) may keep these files at the top level of `templates/` or in a folder
named after their purpose. They still carry the gate.

### Crossplane resources

Cloud resources are created through Crossplane rather than by the agent, so they
follow the deployment instead of being provisioned ahead of it. They are ordinary
templates under the cloud folder and share a shape:

```yaml
{{- if eq .Values.global.cloudProvider "aws" }}
apiVersion: iam.aws.upbound.io/v1beta1
kind: Role
metadata:
  name: {{ .Release.Name }}
  annotations:
    crossplane.io/external-name: "{{ .Release.Name }}"
    argocd.argoproj.io/sync-options: SkipDryRunOnMissingResource=true
    argocd.argoproj.io/sync-wave: '-2'
spec:
  forProvider:
    tags:
      created-by: entigo-infralib
    assumeRolePolicy: |
      ... {{ .Values.global.aws.account }} ... {{ .Values.global.aws.clusterOIDC }} ...
  providerConfigRef:
    name: {{ .Values.global.providerConfigRefName }}
{{- end }}
```

The recurring pieces:

- **`providerConfigRef.name: {{ .Values.global.providerConfigRefName }}`** —
  never hardcode `crossplane-aws` or `crossplane-google`, it comes from
  `values-<cloud>.yaml`.
- **`SkipDryRunOnMissingResource=true`** — the Crossplane CRDs may not exist yet
  the first time ArgoCD syncs the module.
- **A negative `sync-wave`** so the cloud resources are created before the
  workload that needs them. `-3` for activation policies, `-2` for the resources
  themselves is the usual split.
- **`crossplane.io/external-name`** where the cloud resource must have a
  predictable name.
- **`created-by: entigo-infralib`** in tags or labels, so the resources are
  identifiable in the cloud account.
- **A `ManagedResourceActivationPolicy`** (`mrap.yaml`) listing the managed
  resource kinds the module uses, at sync-wave `-3`.

The values these templates read — account, project, OIDC issuer, KMS key —
always come from the agent inputs, never from a value file:

```yaml
global:
  aws:
    account: "{{ .toutput.eks.account }}"
    clusterOIDC: "{{ .toutput.eks.oidc_provider }}"
```

## Updating the upstream chart

1. Bump the version under `dependencies:` in `Chart.yaml`.
2. Run `helm dependency update` in the module folder. This downloads the new
   `charts/*.tgz` and rewrites `Chart.lock`.
3. Commit `Chart.yaml`, `Chart.lock` and the new tarball, and delete the old
   one. The tarball is vendored on purpose: ArgoCD must be able to render the
   chart from this repository alone.
4. Diff the upstream values against ours. New required keys have to be added to
   `values.yaml`, renamed keys have to be followed in `values.yaml`,
   `values-<cloud>.yaml` and the agent inputs.
5. Run `./test.sh` in the module folder.

`version:` and `appVersion:` of the wrapper chart are not tied to the upstream
chart and are not bumped as part of a dependency update. What gets deployed is
decided by the infralib release tag, which the agent writes into the
`Application` as `targetRevision`.

Most of this is done automatically: a bot opens `chore(deps): update ...` pull
requests on `auto-k8s-<module>` branches, changing only the dependency version,
`Chart.lock` and the tarball. Steps 4 and 5 are the parts a human still has to
do.

Charts whose CRDs we manage ourselves — `aws-alb` is the current example — have
extra steps documented in the module's own README, because Helm does not install
subchart `crds/` and ArgoCD renders with `helm template`, which skips `crds/`
entirely. Those CRDs are copied into `templates/` instead.

## Tests

`./test.sh` in a module folder runs static tests, then unit tests if the module
has any. Both also run in CI for every module touched by a pull request.

### Static tests

`helm lint --strict`, `helm template` and `kube-score`, run against the chart
**with `values.yaml` only**. No `values-<cloud>.yaml`, no agent inputs, no
platform config.

That is deliberate — it is the one render that has to work with nothing but what
this repository ships. It also means `global.cloudProvider` is empty, so nothing
under `templates/aws/` or `templates/google/` is exercised. Cloud specific
templates are covered by the unit tests, not here.

#### test/static_values.yaml

Some charts cannot render at all without a value that only ever arrives from the
agent — a cluster name, a bucket, a storage schema. `helm template` fails, and
the module has no way to pass a static test.

For those, `test/static_values.yaml` supplies the minimum placeholder set, and
the static tests add it with `-f`:

```yaml
# aws-alb/test/static_values.yaml
aws-load-balancer-controller:
  clusterName: test
```

Keep it as small as it can be. It exists to make rendering possible, not to
model a realistic installation — anything you can give a sensible default to
belongs in `values.yaml` instead. Five modules currently need one.

### Test environment inputs

`test/<cloud>_<env>.yaml` — `aws_pri.yaml`, `aws_biz.yaml`, `google_pri.yaml`,
`google_biz.yaml` — are the module's `inputs:` for the test platform of that
cloud and environment. CI copies each one into the generated agent config as
that module's input file, so they are agent inputs and support the same
replacement tags.

**A module is only deployed into a test environment if the matching file
exists.** An empty file is fine and means "deploy with defaults"; no file means
"do not test this module there". That is how single cloud modules stay out of
the other cloud's clusters.

Test environments are short lived and reachable without a VPN, so it is normal
and intentional for a test file to point an otherwise internal service at the
external gateway:

```yaml
global:
  internalGateway: external
```

Do not treat that as a mistake to clean up.

`test/templates/` holds throwaway manifests that a test applies to a cluster —
a backend deployment and service to route to, for example. They are not part of
the chart.

### Unit tests

`test/k8s_unit_basic_test.go` runs with terratest against the live test clusters
after the agent has deployed the module, one test function per cloud and
environment. This is where cloud specific templates actually get verified.

## Adding a new module

1. `Chart.yaml` with the upstream dependency, then `helm dependency update`.
2. `values.yaml` with every key the templates read, `values-<cloud>.yaml` with
   `global.cloudProvider` and the per cloud differences.
3. `agent_input.yaml` for the wiring, chained across clouds; per cloud input
   files only for what exists on one cloud alone.
4. `templates/` for the resources we own, cloud specific ones under
   `templates/<cloud>/` and gated on `global.cloudProvider`.
5. `argo-apps.yaml` if the module needs sync options, ignored differences or
   namespace labels.
6. `test.sh` copied from an existing module — it is the same three lines
   everywhere.
7. `test/<cloud>_<env>.yaml` for each environment the module should be deployed
   into, and `test/k8s_unit_basic_test.go` to assert it came up.
8. `test/static_values.yaml` only if the chart cannot render without it.
9. Add the module to `test_k8s()` in `common/generate_config.sh`.
10. Run `./test.sh` before opening the pull request.
