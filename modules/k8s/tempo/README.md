# tempo

Wraps the upstream `tempo` chart (Grafana Tempo, single-binary/monolithic mode) from
`grafana-community/helm-charts`, following the same S3-backed-storage-via-Crossplane-IAM
shape as the `loki` and `mimir` modules.

## Before merging

Everything in this module was written from the `loki`/`mimir` module pattern plus
general knowledge of Tempo's chart shape — it was **not** checked against the real
upstream `values.yaml` (fetching it from the authoring session was blocked). Before
opening a PR:

1. `helm repo add grafana-community https://grafana-community.github.io/helm-charts`
2. `helm search repo grafana-community/tempo --versions` — confirm the version pinned
   in `Chart.yaml` is current.
3. `helm show values grafana-community/tempo --version <x>` — diff against
   `values.yaml` / `values-aws.yaml` / `agent_input*.yaml` here. Everything marked
   `VERIFY` in those files needs a real key-path check, especially:
   - `tempo.storage.trace.*` (S3 backend config)
   - `tempo.serviceAccount.*` (name/create — must match `templates/aws/serviceAccount.yaml`'s
     `metadata.name: tempo` and the IRSA trust policy's `sub` claim in `templates/aws/role.yaml`)
   - `tempo.distributor.receivers.otlp.protocols.*` (OTLP ports)
   - `tempo.image.*` (registry override path, for the ecr-proxy chain in `agent_input.yaml`)
4. `helm dependency update` in this folder — fetches the real `charts/tempo-<version>.tgz`
   and writes `Chart.lock`. Commit both.
5. `./test.sh`.

## Design choices worth revisiting

- **No external route/ingress.** Unlike `loki` and `grafana`, this module does not ship
  an ALB `TargetGroupConfiguration`/`NetworkPolicy` pair or a Gateway API `HTTPRoute`.
  Grafana queries Tempo in-cluster (`http://tempo.tempo:3100`), and nothing outside the
  cluster needs to reach it directly. Add one back (copy from `loki/templates/aws/`) if
  that assumption turns out to be wrong.
- **No trace ingestion wired up yet.** This module only gives you a place to store and
  query traces — nothing in the platform currently emits OTLP. The `alloy` module only
  runs a `logs:` alias today (tailing pods, forwarding to Loki); getting real traces in
  means adding a `traces:` alias there with an OTLP receiver/exporter pair pointed at
  this module's service, and adding an application-side OTLP SDK/exporter.
- **No Grafana datasource yet.** `grafana`'s `values.yaml`/`agent_input.yaml` declare
  `global.datasources.loki.hostname` / `.prometheus.hostname`, consumed by the
  datasources `ConfigMap` in `grafana/templates/aws/configMap.yaml`. Add a matching
  `global.datasources.tempo.hostname` and a `type: tempo` entry there once this module
  is deployed.
- **14 day trace retention** (`tempo.compactor.compaction.block_retention`) is a cheap
  starting default, not a researched number — tune it once you know real trace volume
  and the S3 storage cost that comes with it.
